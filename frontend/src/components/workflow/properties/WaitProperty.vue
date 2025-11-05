<template>
  <div class="basic-section">
    <h5>⏱️ 等待延时</h5>
    <div class="form-item">
      <label>等待时间(ms)</label>
      <input v-model.number="formData.duration" type="number" placeholder="1000" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ formData: any }>()
</script>

<style scoped>
@import './common.css';
</style>
