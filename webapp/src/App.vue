<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-mark" aria-hidden="true">
          <span class="brand-initials">NP</span>
          <svg class="brand-pulse" viewBox="0 0 32 16" role="presentation" aria-hidden="true">
            <path
              d="M1 8H7L10 3L14 13L18 8H31"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            ></path>
          </svg>
        </div>
        <div class="brand-text">
          <div class="brand-title">NginxPulse</div>
          <div class="brand-sub">Nginx 访问分析</div>
        </div>
      </div>
      <nav class="menu">
        <RouterLink to="/" class="menu-item" :class="{ active: isActive('/') }">概况</RouterLink>
        <RouterLink to="/daily" class="menu-item" :class="{ active: isActive('/daily') }">数据日报</RouterLink>
        <RouterLink to="/realtime" class="menu-item" :class="{ active: isActive('/realtime') }">实时</RouterLink>
        <RouterLink to="/logs" class="menu-item" :class="{ active: isActive('/logs') }">访问明细</RouterLink>
      </nav>
      <div class="sidebar-footer">
        <template v-if="isActive('/')">
          <div class="sidebar-label">近期活跃</div>
          <div class="sidebar-metric">
            <div class="sidebar-metric-value">{{ liveVisitorText }}</div>
            <div class="sidebar-metric-label">15 分钟活跃访客</div>
          </div>
        </template>
        <template v-else>
          <div class="sidebar-label">{{ sidebarLabel }}</div>
          <div class="sidebar-hint">{{ sidebarHint }}</div>
        </template>
      </div>
    </aside>

    <main class="main-content" :class="[mainClass, { 'parsing-lock': parsingActive }]">
      <div v-if="demoMode" class="demo-mode-banner">
        <span class="demo-mode-badge">演示模式</span>
        <span class="demo-mode-text">
          当前处于演示模式，数据均为模拟数据。项目源码请移步：
          <a href="https://github.com/likaia/nginxpulse/" target="_blank" rel="noopener">https://github.com/likaia/nginxpulse/</a>
        </span>
      </div>
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, provide, ref, watch } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { fetchAppStatus } from '@/api';

const route = useRoute();

const sidebarLabel = computed(() => (route.meta.sidebarLabel as string) || '');
const sidebarHint = computed(() => (route.meta.sidebarHint as string) || '');
const mainClass = computed(() => (route.meta.mainClass as string) || '');

const isActive = (path: string) => route.path === path;

const isDark = ref(localStorage.getItem('darkMode') === 'true');
const parsingActive = ref(false);
const liveVisitorCount = ref<number | null>(null);
const demoMode = ref(false);

const applyTheme = (value: boolean) => {
  if (value) {
    document.body.classList.add('dark-mode');
    document.documentElement.classList.add('dark-mode');
    localStorage.setItem('darkMode', 'true');
  } else {
    document.body.classList.remove('dark-mode');
    document.documentElement.classList.remove('dark-mode');
    localStorage.setItem('darkMode', 'false');
  }
};

const toggleTheme = () => {
  isDark.value = !isDark.value;
};

onMounted(() => {
  applyTheme(isDark.value);
  refreshAppStatus();
});

watch(isDark, (value) => {
  applyTheme(value);
});

provide('theme', {
  isDark,
  toggle: toggleTheme,
});

provide('setParsingActive', (value: boolean) => {
  parsingActive.value = value;
});

provide('setLiveVisitorCount', (value: number | null) => {
  liveVisitorCount.value = value;
});

provide('demoMode', demoMode);

async function refreshAppStatus() {
  try {
    const status = await fetchAppStatus();
    demoMode.value = Boolean(status.demo_mode);
  } catch (error) {
    console.error('获取系统状态失败:', error);
  }
}

const liveVisitorText = computed(() =>
  Number.isFinite(liveVisitorCount.value ?? NaN)
    ? (liveVisitorCount.value as number).toLocaleString('zh-CN')
    : '--'
);
</script>

<style lang="scss" scoped>
.demo-mode-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
  border-radius: 14px;
  border: 1px solid rgba(239, 68, 68, 0.2);
  background: rgba(239, 68, 68, 0.08);
  color: #991b1b;
  font-size: 13px;
  font-weight: 500;
  box-shadow: var(--shadow-soft);
}

.demo-mode-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(239, 68, 68, 0.14);
  color: #b91c1c;
  font-weight: 700;
  font-size: 12px;
  letter-spacing: 0.4px;
}

.demo-mode-text {
  color: inherit;
  line-height: 1.5;
}

.demo-mode-text a {
  color: inherit;
  text-decoration: underline;
  text-underline-offset: 3px;
}
</style>
