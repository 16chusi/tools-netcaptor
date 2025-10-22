<template>
  <div v-if="visible" class="drawer-overlay" @click="$emit('close')">
    <div class="drawer" @click.stop>
      <div class="drawer-header">
        <h3>🔒 NetCaptor HTTPS 证书</h3>
        <button @click="$emit('close')" class="close-icon">✕</button>
      </div>
      <div class="drawer-content">
        <p>要捕获 HTTPS 请求，需要安装并信任此 CA 证书。</p>
        
        <div class="cert-path">
          <strong>证书位置:</strong>
          <code>{{ certPath }}</code>
        </div>

        <div class="install-steps">
          <h4>Windows 安装步骤</h4>
          <ol>
            <li>双击 netcaptor-ca.crt 文件</li>
            <li>点击"安装证书"</li>
            <li>选择"当前用户"</li>
            <li>选择"将所有证书放入下列存储"</li>
            <li>浏览并选择"受信任的根证书颁发机构"</li>
            <li>完成安装并重启浏览器</li>
          </ol>

          <h4>macOS 安装步骤</h4>
          <ol>
            <li>双击 netcaptor-ca.crt 文件</li>
            <li>在钥匙串访问中找到 "NetCaptor CA"</li>
            <li>双击证书，展开"信任"</li>
            <li>选择"始终信任"</li>
            <li>重启浏览器</li>
          </ol>

          <h4>Linux/Ubuntu 安装步骤</h4>
          
          <p><strong>方法1: 系统级安装（推荐）</strong></p>
          <pre>sudo cp {{ certPath }} /usr/local/share/ca-certificates/netcaptor.crt
sudo update-ca-certificates
# 重启浏览器</pre>

          <p><strong>方法2: Chrome/Chromium 专用</strong></p>
          <pre>mkdir -p $HOME/.pki/nssdb
certutil -d sql:$HOME/.pki/nssdb -A -t "C,," -n "GoProxy CA" -i {{ certPath }}
# 如果没有certutil，先安装: sudo apt install libnss3-tools
# 重启浏览器</pre>

          <p><strong>方法3: 临时测试（不推荐）</strong></p>
          <p>启动Chrome时添加参数忽略证书错误：</p>
          <pre>google-chrome --proxy-server="127.0.0.1:8888" --ignore-certificate-errors</pre>

          <h4>Chrome 快捷方式</h4>
          <ol>
            <li>打开 Chrome 设置</li>
            <li>搜索"证书"</li>
            <li>点击"管理证书"</li>
            <li>选择"受信任的根证书颁发机构"</li>
            <li>点击"导入"，选择 netcaptor-ca.crt</li>
            <li>重启浏览器</li>
          </ol>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean
  certPath: string
}>()

defineEmits<{
  close: []
}>()
</script>

<style scoped>
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
}

.drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 600px;
  background: white;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  background: #f8f9fa;
}

.drawer-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.close-icon {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: #5f6368;
  cursor: pointer;
  font-size: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.2s;
}

.close-icon:hover {
  background: #e8eaed;
}

.drawer-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  text-align: left;
}

.drawer-content > p {
  color: #666;
  margin-bottom: 16px;
  line-height: 1.6;
  text-align: left;
}

.cert-path {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 16px;
  text-align: left;
}

.cert-path strong {
  display: block;
  margin-bottom: 8px;
}

.cert-path code {
  display: block;
  padding: 8px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  font-size: 12px;
  word-break: break-all;
  text-align: left;
}

.install-steps {
  text-align: left;
}

.install-steps h4 {
  margin: 24px 0 12px 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
  text-align: left;
}

.install-steps ol {
  margin: 0 0 20px 0;
  padding-left: 24px;
  font-size: 13px;
  line-height: 1.8;
  color: #333;
  text-align: left;
}

.install-steps ol li {
  margin-bottom: 6px;
  text-align: left;
}

.install-steps p {
  margin: 12px 0 8px 0;
  color: #666;
  line-height: 1.6;
  text-align: left;
}

.install-steps pre {
  background: #2d2d30;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 4px;
  font-size: 11px;
  overflow-x: auto;
  margin: 8px 0 16px 0;
  line-height: 1.5;
  text-align: left;
}
</style>
