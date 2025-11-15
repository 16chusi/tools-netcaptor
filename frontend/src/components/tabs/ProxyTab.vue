<template>
  <div class="proxy-tab">
    <div class="setting-item">
      <label>启用网络代理</label>
      <div class="toggle-switch">
        <input 
          type="checkbox" 
          id="proxy-enabled" 
          v-model="proxyConfig.enabled"
          @change="saveConfig"
        />
        <label for="proxy-enabled" class="switch-label"></label>
      </div>
      <small>启用后，所有网络请求将通过配置的代理服务器</small>
    </div>

    <div v-if="proxyConfig.enabled" class="proxy-settings">
      <div class="setting-item">
        <label>代理类型</label>
        <select v-model="proxyConfig.type" @change="saveConfig" class="setting-select">
          <option value="http">HTTP</option>
          <option value="https">HTTPS</option>
          <option value="socks5">SOCKS5</option>
        </select>
      </div>

      <div class="setting-item">
        <label>代理服务器地址</label>
        <input 
          v-model="proxyConfig.host" 
          @input="saveConfig"
          type="text" 
          class="setting-input" 
          placeholder="例如: 127.0.0.1 或 proxy.example.com"
        />
      </div>

      <div class="setting-item">
        <label>端口</label>
        <input 
          v-model.number="proxyConfig.port" 
          @input="saveConfig"
          type="number" 
          class="setting-input" 
          placeholder="例如: 8080"
          min="1"
          max="65535"
        />
      </div>

      <div class="setting-item">
        <label>认证</label>
        <div class="toggle-switch">
          <input 
            type="checkbox" 
            id="proxy-auth" 
            v-model="proxyConfig.auth.enabled"
            @change="saveConfig"
          />
          <label for="proxy-auth" class="switch-label"></label>
        </div>
        <small>如果代理服务器需要用户名和密码</small>
      </div>

      <div v-if="proxyConfig.auth.enabled" class="auth-settings">
        <div class="setting-item">
          <label>用户名</label>
          <input 
            v-model="proxyConfig.auth.username" 
            @input="saveConfig"
            type="text" 
            class="setting-input" 
            placeholder="代理用户名"
          />
        </div>

        <div class="setting-item">
          <label>密码</label>
          <input 
            v-model="proxyConfig.auth.password" 
            @input="saveConfig"
            type="password" 
            class="setting-input" 
            placeholder="代理密码"
          />
        </div>
      </div>

      <div class="setting-item">
        <label>绕过代理的地址</label>
        <textarea 
          v-model="proxyConfig.bypass" 
          @input="saveConfig"
          class="setting-textarea" 
          placeholder="每行一个地址，支持通配符&#10;例如:&#10;localhost&#10;127.0.0.1&#10;*.local&#10;192.168.*"
          rows="4"
        ></textarea>
        <small>这些地址将直接连接，不通过代理</small>
      </div>

      <div class="setting-item">
        <label>连接测试</label>
        <div class="test-section">
          <input 
            v-model="testUrl" 
            type="text" 
            class="test-url-input" 
            placeholder="输入要测试的URL"
          />
          <button @click="testConnection" :disabled="testing" class="test-btn">
            {{ testing ? '测试中...' : '测试连接' }}
          </button>
        </div>
        <div v-if="testResult" :class="['test-result', testResult.success ? 'success' : 'error']">
          {{ testResult.message }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {proxy} from "../../../wailsjs/go/models";
import ProxyConfig = proxy.ProxyConfig;
import ProxyAuth = proxy.ProxyAuth;
import {GetProxyConfig, SetProxyConfig, TestProxyConnectionWithURL} from "../../../wailsjs/go/network/NetworkApp";

const proxyConfig = ref< ProxyConfig>(new  ProxyConfig({
  enabled: false,
  type: 'http',
  host: '',
  port: 8080,
  auth: new  ProxyAuth({
    enabled: false,
    username: '',
    password: ''
  }),
  bypass: 'localhost\n127.0.0.1\n*.local'
}))

const testing = ref(false)
const testResult = ref<{ success: boolean; message: string } | null>(null)
const testUrl = ref('https://www.google.com')

onMounted(() => {
  loadConfig()
})

async function loadConfig() {
  try {
    const config = await GetProxyConfig()
    if (config) {
      proxyConfig.value = config
    }
  } catch (e) {
    console.error('加载代理配置失败:', e)
  }
}

async function saveConfig() {
  try {
    await SetProxyConfig(proxyConfig.value)
    testResult.value = null // 清除之前的测试结果
  } catch (e) {
    console.error('保存代理配置失败:', e)
    alert('保存配置失败: ' + e)
  }
}

async function testConnection() {
  if (!proxyConfig.value.host || !proxyConfig.value.port) {
    testResult.value = { success: false, message: '请先配置代理服务器地址和端口' }
    return
  }

  if (!testUrl.value.trim()) {
    testResult.value = { success: false, message: '请输入要测试的URL' }
    return
  }

  testing.value = true
  testResult.value = null

  try {
    const result = await TestProxyConnectionWithURL(testUrl.value.trim())
    testResult.value = { success: result.success, message: result.message }
  } catch (error) {
    testResult.value = { success: false, message: '测试连接时发生错误: ' + error }
  } finally {
    testing.value = false
  }
}
</script>

<style scoped>
.proxy-tab {
  padding: 0;
}

.setting-item {
  margin-bottom: 24px;
}

.setting-item label {
  display: block;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.setting-item small {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: #666;
}

.setting-input,
.setting-select,
.setting-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 4px;
  font-size: 13px;
  box-sizing: border-box;
}

.setting-input,
.setting-select {
  height: 36px;
}

.setting-textarea {
  resize: vertical;
  font-family: 'Courier New', monospace;
  line-height: 1.4;
}

.setting-input:focus,
.setting-select:focus,
.setting-textarea:focus {
  outline: none;
  border-color: #1a73e8;
  box-shadow: 0 0 0 2px rgba(26, 115, 232, 0.1);
}

.toggle-switch {
  position: relative;
  display: inline-block;
}

.toggle-switch input[type="checkbox"] {
  opacity: 0;
  width: 0;
  height: 0;
}

.switch-label {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
  background-color: #ccc;
  border-radius: 24px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.switch-label::after {
  content: '';
  position: absolute;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background-color: white;
  top: 2px;
  left: 2px;
  transition: transform 0.3s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.toggle-switch input:checked + .switch-label {
  background-color: #1a73e8;
}

.toggle-switch input:checked + .switch-label::after {
  transform: translateX(20px);
}

.proxy-settings {
  margin-left: 16px;
  padding-left: 16px;
  border-left: 2px solid #e0e0e0;
}

.auth-settings {
  margin-left: 16px;
  padding-left: 16px;
  border-left: 2px solid #e0e0e0;
  background: #f8f9fa;
  border-radius: 4px;
  padding: 16px;
  margin-top: 12px;
}

.test-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.test-url-input {
  width: 100%;
  height: 36px;
  padding: 0 12px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 4px;
  font-size: 13px;
  box-sizing: border-box;
}

.test-url-input:focus {
  outline: none;
  border-color: #1a73e8;
  box-shadow: 0 0 0 2px rgba(26, 115, 232, 0.1);
}

.test-btn {
  padding: 8px 16px;
  border: 1px solid #1a73e8;
  background: #1a73e8;
  color: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.test-btn:hover:not(:disabled) {
  background: #1557b0;
  border-color: #1557b0;
}

.test-btn:disabled {
  background: #dadce0;
  border-color: #dadce0;
  cursor: not-allowed;
}

.test-result {
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.test-result.success {
  background: #e8f5e8;
  color: #2e7d32;
  border: 1px solid #c8e6c9;
}

.test-result.error {
  background: #ffebee;
  color: #c62828;
  border: 1px solid #ffcdd2;
}
</style>
