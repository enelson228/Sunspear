<template>
  <div class="page">
    <h1>APP DETAIL</h1>
    <p class="text-secondary">App ID: {{ $route.params.id }}</p>
  </div>
</template>

<style scoped>
.page {
  padding: var(--space-2xl);
}
</style>
