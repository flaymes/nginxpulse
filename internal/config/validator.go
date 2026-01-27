package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidateOptions struct {
	CheckPaths  bool
	CheckRemote bool
}

type ValidationResult struct {
	Errors   []FieldError `json:"errors"`
	Warnings []FieldError `json:"warnings"`
}

func ValidateConfig(cfg *Config, opts ValidateOptions) ValidationResult {
	result := ValidationResult{}
	addError := func(field, msg string) {
		result.Errors = append(result.Errors, FieldError{Field: field, Message: msg})
	}
	addWarning := func(field, msg string) {
		result.Warnings = append(result.Warnings, FieldError{Field: field, Message: msg})
	}

	if cfg == nil {
		addError("config", "配置不能为空")
		return result
	}

	if len(cfg.Websites) == 0 {
		addError("websites", "至少需要配置一个站点")
	}

	allowMissingPaths := IsSetupMode()
	for i, site := range cfg.Websites {
		sitePrefix := fmt.Sprintf("websites[%d]", i)
		if strings.TrimSpace(site.Name) == "" {
			addError(sitePrefix+".name", "站点名称不能为空")
		}

		if len(site.Sources) == 0 {
			if strings.TrimSpace(site.LogPath) == "" {
				addError(sitePrefix+".logPath", "日志路径不能为空")
			} else if opts.CheckPaths {
				if err, warn := validatePath(site.LogPath, allowMissingPaths); err != nil {
					if warn {
						addWarning(sitePrefix+".logPath", err.Error())
					} else {
						addError(sitePrefix+".logPath", err.Error())
					}
				}
			}
			continue
		}

		seen := map[string]struct{}{}
		for sidx, src := range site.Sources {
			srcPrefix := fmt.Sprintf("%s.sources[%d]", sitePrefix, sidx)
			id := strings.TrimSpace(src.ID)
			if id == "" {
				addError(srcPrefix+".id", "source.id 不能为空")
			} else if _, ok := seen[id]; ok {
				addError(srcPrefix+".id", "source.id 重复")
			} else {
				seen[id] = struct{}{}
			}

			stype := strings.ToLower(strings.TrimSpace(src.Type))
			if stype == "" {
				addError(srcPrefix+".type", "source.type 不能为空")
				continue
			}

			switch stype {
			case "local":
				if strings.TrimSpace(src.Path) == "" && strings.TrimSpace(src.Pattern) == "" {
					addError(srcPrefix, "local 需要 path 或 pattern")
				} else if opts.CheckPaths {
					if src.Path != "" {
						if err, warn := validatePath(src.Path, allowMissingPaths); err != nil {
							if warn {
								addWarning(srcPrefix+".path", err.Error())
							} else {
								addError(srcPrefix+".path", err.Error())
							}
						}
					}
					if src.Pattern != "" {
						if err, warn := validatePath(src.Pattern, allowMissingPaths); err != nil {
							if warn {
								addWarning(srcPrefix+".pattern", err.Error())
							} else {
								addError(srcPrefix+".pattern", err.Error())
							}
						}
					}
				}
			case "sftp":
				if strings.TrimSpace(src.Host) == "" {
					addError(srcPrefix+".host", "sftp.host 不能为空")
				}
				if strings.TrimSpace(src.User) == "" {
					addError(srcPrefix+".user", "sftp.user 不能为空")
				}
				if src.Auth == nil || (strings.TrimSpace(src.Auth.KeyFile) == "" && strings.TrimSpace(src.Auth.Password) == "") {
					addError(srcPrefix+".auth", "sftp 需要 keyFile 或 password")
				}
				if strings.TrimSpace(src.Path) == "" && strings.TrimSpace(src.Pattern) == "" {
					addError(srcPrefix, "sftp 需要 path 或 pattern")
				} else if opts.CheckRemote {
					addWarning(srcPrefix, "远端路径校验会在后续版本支持")
				}
			case "http":
				if strings.TrimSpace(src.URL) == "" {
					addError(srcPrefix+".url", "http.url 不能为空")
				}
				if src.Index != nil && strings.TrimSpace(src.Index.URL) == "" {
					addError(srcPrefix+".index.url", "http.index.url 不能为空")
				}
			case "s3":
				if strings.TrimSpace(src.Bucket) == "" {
					addError(srcPrefix+".bucket", "s3.bucket 不能为空")
				}
				if (strings.TrimSpace(src.AccessKey) == "") != (strings.TrimSpace(src.SecretKey) == "") {
					addError(srcPrefix+".accessKey", "s3 accessKey/secretKey 需同时配置")
				}
			case "agent":
				// no-op
			default:
				addError(srcPrefix+".type", "不支持的 source.type")
			}
		}
	}

	if strings.TrimSpace(cfg.Database.Driver) == "" {
		addError("database.driver", "数据库驱动不能为空")
	} else if strings.TrimSpace(cfg.Database.Driver) != "postgres" {
		addError("database.driver", "仅支持 postgres 驱动")
	}
	if strings.TrimSpace(cfg.Database.DSN) == "" {
		addError("database.dsn", "数据库 DSN 不能为空")
	}
	if cfg.System.LogRetentionDays <= 0 {
		addError("system.logRetentionDays", "logRetentionDays 必须大于 0")
	}
	if cfg.System.ParseBatchSize <= 0 {
		addError("system.parseBatchSize", "parseBatchSize 必须大于 0")
	}
	if cfg.System.IPGeoCacheLimit <= 0 {
		addError("system.ipGeoCacheLimit", "ipGeoCacheLimit 必须大于 0")
	}

	if len(cfg.PVFilter.StatusCodeInclude) == 0 {
		addError("pvFilter.statusCodeInclude", "statusCodeInclude 不能为空")
	}
	if len(cfg.PVFilter.ExcludePatterns) == 0 {
		addError("pvFilter.excludePatterns", "excludePatterns 不能为空")
	}

	return result
}

func validatePath(value string, allowMissing bool) (error, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("路径不能为空"), false
	}
	if strings.Contains(value, "*") {
		matches, err := filepath.Glob(value)
		if err != nil || len(matches) == 0 {
			if allowMissing {
				return fmt.Errorf("日志路径未匹配到任何文件（可先完成配置，稍后再创建）"), true
			}
			return fmt.Errorf("日志路径未匹配到任何文件"), false
		}
		return nil, false
	}
	if _, err := os.Stat(value); err != nil {
		if os.IsNotExist(err) && allowMissing {
			return fmt.Errorf("日志路径不存在或不可访问（可先完成配置，稍后再创建）"), true
		}
		if os.IsPermission(err) {
			return fmt.Errorf("日志路径不可访问，请确保容器运行用户有权限"), false
		}
		return fmt.Errorf("日志路径不存在或不可访问"), false
	}
	return nil, false
}
