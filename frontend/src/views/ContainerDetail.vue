<template>
  <div class="page">
    <h1>CONTAINER DETAIL</h1>
    <p class="text-secondary">Container ID: {{ $route.params.id }}</p>
  </div>
</template>

<style scoped>
.page {
  padding: var(--space-2xl);
}
</style>
