<template>
  <nav class="navbar glass">
    <div class="container">
      <div class="navbar-content">
        <div class="logo-section">
          <div class="logo-icon">◆</div>
          <span class="logo-text">SUNSPEAR</span>
        </div>

        <div :class="['nav-links', { 'mobile-open': mobileMenuOpen }]">
          <router-link to="/" class="nav-link" @click="closeMobileMenu">Dashboard</router-link>
          <router-link to="/containers" class="nav-link" @click="closeMobileMenu">Containers</router-link>
          <router-link to="/images" class="nav-link" @click="closeMobileMenu">Images</router-link>
          <router-link to="/volumes" class="nav-link" @click="closeMobileMenu">Volumes</router-link>
          <router-link to="/networks" class="nav-link" @click="closeMobileMenu">Networks</router-link>
          <router-link to="/compose" class="nav-link" @click="closeMobileMenu">Compose</router-link>
          <router-link to="/apps" class="nav-link" @click="closeMobileMenu">App Store</router-link>
          <router-link to="/system" class="nav-link" @click="closeMobileMenu">System</router-link>
          <router-link to="/settings" class="nav-link" @click="closeMobileMenu">Settings</router-link>
          <button class="nav-link mobile-logout" @click="handleLogout">LOGOUT</button>
        </div>

        <div class="nav-actions">
          <button class="logout-btn" @click="handleLogout">LOGOUT</button>
        </div>

        <button class="mobile-menu-btn" @click="toggleMobileMenu" aria-label="Toggle menu">
          <div :class="['hamburger', { open: mobileMenuOpen }]">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </button>
      </div>
    </div>
    <div v-if="mobileMenuOpen" class="mobile-overlay" @click="closeMobileMenu"></div>
  </nav>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const mobileMenuOpen = ref(false)

function toggleMobileMenu() {
  mobileMenuOpen.value = !mobileMenuOpen.value
}

function closeMobileMenu() {
  mobileMenuOpen.value = false
}

function handleLogout() {
  closeMobileMenu()
  authStore.logout()
  router.push('/login')
}
</script>

<style>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  height: 64px;
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.navbar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.logo-icon {
  color: var(--reach-amber);
  font-size: 1.5rem;
}

.logo-text {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  color: var(--reach-amber);
}

.nav-links {
  display: flex;
  gap: var(--space-xl);
}

.nav-link {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  transition: color var(--transition-base);
  position: relative;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  text-decoration: none;
}

.nav-link:hover {
  color: var(--reach-amber);
}

.nav-link.router-link-active {
  color: var(--reach-amber);
}

.nav-link.router-link-active::after {
  content: '';
  position: absolute;
  bottom: -20px;
  left: 0;
  right: 0;
  height: 2px;
  background-color: var(--reach-amber);
}

.nav-actions {
  display: flex;
  align-items: center;
}

.logout-btn {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
  background: transparent;
  border: 1px solid rgba(74, 85, 104, 0.5);
  padding: var(--space-xs) var(--space-md);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-base);
}

.logout-btn:hover {
  color: var(--reach-amber);
  border-color: var(--reach-amber);
}

.mobile-logout {
  display: none;
}

.mobile-menu-btn {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: var(--space-sm);
}

.hamburger {
  display: flex;
  flex-direction: column;
  gap: 5px;
  width: 24px;
}

.hamburger span {
  display: block;
  height: 2px;
  background-color: var(--text-secondary);
  transition: all 0.3s ease;
}

.hamburger.open span:nth-child(1) {
  transform: rotate(45deg) translate(5px, 5px);
}

.hamburger.open span:nth-child(2) {
  opacity: 0;
}

.hamburger.open span:nth-child(3) {
  transform: rotate(-45deg) translate(5px, -5px);
}

.mobile-overlay {
  display: none;
}

@media (max-width: 768px) {
  .nav-links {
    display: none;
    position: absolute;
    top: 64px;
    left: 0;
    right: 0;
    flex-direction: column;
    background-color: var(--reach-steel);
    border-bottom: 1px solid rgba(74, 85, 104, 0.3);
    padding: var(--space-md);
    gap: 0;
    z-index: 100;
  }

  .nav-links.mobile-open {
    display: flex;
  }

  .nav-links .nav-link {
    padding: var(--space-md) var(--space-lg);
    border-bottom: 1px solid rgba(74, 85, 104, 0.2);
    width: 100%;
    text-align: left;
  }

  .nav-link.router-link-active::after {
    display: none;
  }

  .nav-actions {
    display: none;
  }

  .mobile-logout {
    display: block;
    color: var(--reach-orange);
    margin-top: var(--space-sm);
  }

  .mobile-menu-btn {
    display: flex;
    align-items: center;
  }

  .mobile-overlay {
    display: block;
    position: fixed;
    top: 64px;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 99;
  }
}
</style>
