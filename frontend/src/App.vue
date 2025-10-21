<script lang="ts" setup>
import NetworkCapture from './components/NetworkCapture.vue'
</script>

<template>
  <div class="app-wrapper">
    <NetworkCapture />
  </div>
</template>

<style scoped>
.app-wrapper {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
</style>
