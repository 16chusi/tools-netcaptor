<template>
  <div class="data-section">
    <h5>🔐 数据解密</h5>
    <div class="form-item">
      <label>数据来源</label>
      <input v-model="formData.dataVariable" placeholder="变量名，如 data" />
      <div class="variable-hint">输入变量名，不需要 {}</div>
    </div>
    <div class="form-item">
      <label>解密算法</label>
      <select v-model="formData.algorithm">
        <option value="sm4-ecb">SM4-ECB</option>
        <option value="sm4-cbc">SM4-CBC</option>
      </select>
    </div>
    <div class="form-item">
      <label>密钥 (Hex)</label>
      <input v-model="formData.key" placeholder="46696e32416e63304571753245727934" />
      <div class="variable-hint">16进制格式的密钥</div>
    </div>
    <div class="form-item" v-if="formData.algorithm === 'sm4-cbc'">
      <label>IV向量 (Hex)</label>
      <input v-model="formData.iv" placeholder="可选，留空使用零向量" />
      <div class="variable-hint">16进制格式的初始化向量</div>
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="decryptedData" />
      <div class="variable-hint">结果将保存到: {{ '{' + (formData.saveToVariable || 'decryptedData') + '}' }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ formData: any }>()
</script>

<style scoped>
@import './common.css';
</style>
