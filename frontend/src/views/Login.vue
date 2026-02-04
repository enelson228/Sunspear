<template>
  <div class="login-container">
    <div class="login-box hud-corners">
      <div class="login-header">
        <div class="logo">
          <div class="logo-icon">◆</div>
          <h1>SUNSPEAR</h1>
        </div>
        <p class="subtitle">Docker Management System</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <Input
          v-model="username"
          label="USERNAME"
          type="text"
          placeholder="Enter username"
          :disabled="loading"
        />

        <Input
          v-model="password"
          label="PASSWORD"
          type="password"
          placeholder="Enter password"
          :disabled="loading"
        />

        <div v-if="error" class="error-message">
          {{ error }}
        </div>

        <Button
          type="submit"
          variant="primary"
          :loading="loading"
          :show-accent="true"
          class="w-full mt-lg"
        >
          {{ loading ? 'AUTHENTICATING...' : 'LOGIN' }}
        </Button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  error.value = ''
  loading.value = true

  try {
    await authStore.login(username.value, password.value)
    router.push('/')
  } catch (err) {
    error.value = 'Invalid credentials. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(
    circle at center,
    rgba(246, 166, 35, 0.05) 0%,
    transparent 70%
  );
}

.login-box {
  width: 100%;
  max-width: 400px;
  padding: var(--space-2xl);
  background-color: var(--reach-steel);
  border: 1px solid rgba(246, 166, 35, 0.3);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
}

.login-header {
  text-align: center;
  margin-bottom: var(--space-2xl);
}

.logo {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
}

.logo-icon {
  color: var(--reach-amber);
  font-size: 2rem;
  text-shadow: 0 0 20px var(--reach-amber);
}

.login-header h1 {
  font-family: var(--font-display);
  font-size: 2rem;
  letter-spacing: 0.3em;
  color: var(--reach-amber);
  margin: 0;
}

.subtitle {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
  margin-top: var(--space-sm);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.error-message {
  padding: var(--space-sm) var(--space-md);
  background-color: rgba(232, 93, 4, 0.2);
  border: 1px solid rgba(232, 93, 4, 0.4);
  border-radius: var(--radius-sm);
  color: var(--reach-orange);
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-align: center;
}
</style>
