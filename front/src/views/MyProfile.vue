<script setup>
import { ref, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CheckCircle, User, Phone, Mail, MapPin, Calendar, Pencil, UserPen, ChevronRight, Building2, Wallet, Lock, Bell, LogOut, ArrowLeft, Clock, AlertCircle, Loader2, CloudUpload, Info, FileImage, Camera, ArrowDown, ArrowUp, Key, Smartphone, Shield, MessageCircle, Hammer, Banknote, MessageSquare } from 'lucide-vue-next'
import { getMyCompanyStatus, applyCompany, recognizeLicense, getPhone, getUsername, logout } from '@/api/auth'

const router = useRouter()
const isLoggedIn = inject('isLoggedIn')

const userInfo = ref({
  name: '',
  avatar: '',
  company: '',
  phone: '',
  email: '',
  location: '',
  joinTime: '',
  role: 'supplier',
  verified: false
})

const stats = ref({
  totalBids: 0,
  successRate: 0,
  completedOrders: 0
})

// 企业认证状态
const companyStatus = ref({
  hasVerifiedCompany: false,
  company: null,
  pendingApplication: null,
  latestRejected: null
})

const loading = ref(true)
const submitting = ref(false)

// 当前显示的模块
const currentModule = ref('main')
const slideDirection = ref('left')

// 切换到子模块
const goToModule = (module) => {
  slideDirection.value = 'left'
  currentModule.value = module
}

// 退出登录
const handleLogout = () => {
  logout()
  isLoggedIn.value = false
  router.push('/')
}

// 返回主菜单
const goBack = () => {
  slideDirection.value = 'right'
  currentModule.value = 'main'
}

// 企业认证表单数据
const enterpriseForm = ref({
  companyName: '',
  creditCode: '',
  legalPerson: '',
  registeredCapital: '',
  establishDate: '',
  businessScope: '',
  address: '',
  license: null
})

// OCR识别状态
const ocrStep = ref('upload') // 'upload' | 'recognized'
const recognizing = ref(false)
const imagePath = ref('')

// 文件输入引用
const fileInputRef = ref(null)
const selectedFileName = ref('')

// 处理文件选择并触发OCR识别
const handleFileSelect = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  // 验证文件类型
  const allowedTypes = ['image/jpeg', 'image/png', 'image/jpg', 'application/pdf']
  if (!allowedTypes.includes(file.type)) {
    ElMessage.warning('只支持 JPG、PNG、PDF 格式的文件')
    return
  }
  // 验证文件大小 (5MB)
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning('文件大小不能超过 5MB')
    return
  }

  enterpriseForm.value.license = file
  selectedFileName.value = file.name

  // 上传并进行OCR识别
  recognizing.value = true
  try {
    const result = await recognizeLicense(file)
    // 填充识别结果到表单
    enterpriseForm.value.companyName = result.data.companyName || ''
    enterpriseForm.value.creditCode = result.data.businessLicenseNo || ''
    enterpriseForm.value.legalPerson = result.data.legalPerson || ''
    enterpriseForm.value.registeredCapital = result.data.registeredCapital || ''
    enterpriseForm.value.establishDate = result.data.establishDate || ''
    enterpriseForm.value.businessScope = result.data.businessScope || ''
    enterpriseForm.value.address = result.data.address || ''
    imagePath.value = result.image
    ocrStep.value = 'recognized'
  } catch (error) {
    ElMessage.error(error.message || 'OCR识别失败，请重试')
  } finally {
    recognizing.value = false
  }
}

// 重新上传
const resetUpload = () => {
  ocrStep.value = 'upload'
  enterpriseForm.value = {
    companyName: '', creditCode: '', legalPerson: '',
    registeredCapital: '', establishDate: '', businessScope: '',
    address: '', license: null
  }
  selectedFileName.value = ''
  imagePath.value = ''
}

// 触发文件选择
const triggerFileSelect = () => {
  fileInputRef.value?.click()
}

// 加载企业认证状态
const loadCompanyStatus = async () => {
  try {
    const status = await getMyCompanyStatus()
    companyStatus.value = status

    // 更新 userInfo
    if (status.hasVerifiedCompany && status.company) {
      userInfo.value.verified = true
      userInfo.value.company = status.company.name
    } else {
      userInfo.value.verified = false
      userInfo.value.company = ''
    }
  } catch (error) {
    console.error('获取企业认证状态失败:', error)
  }
}

// 提交企业认证
const submitEnterprise = async () => {
  if (!enterpriseForm.value.companyName) {
    ElMessage.warning('请输入企业名称')
    return
  }
  if (!enterpriseForm.value.creditCode) {
    ElMessage.warning('请输入统一社会信用代码')
    return
  }
  if (!imagePath.value) {
    ElMessage.warning('请先上传营业执照进行识别')
    return
  }

  submitting.value = true
  try {
    await applyCompany({
      name: enterpriseForm.value.companyName,
      businessLicenseNo: enterpriseForm.value.creditCode,
      imagePath: imagePath.value
    })
    ElMessage.success('企业认证申请已提交，请等待审核')

    // 重新加载状态
    await loadCompanyStatus()

    // 清空表单
    resetUpload()

    goBack()
  } catch (error) {
    ElMessage.error(error.message || '提交失败，请重试')
  } finally {
    submitting.value = false
  }
}

// 页面加载时获取数据
onMounted(async () => {
  loading.value = true
  try {
    // 从 localStorage 获取手机号和用户名
    const phone = getPhone()
    if (phone) {
      // 脱敏显示
      userInfo.value.phone = phone.length === 11
        ? phone.slice(0, 3) + '****' + phone.slice(-4)
        : phone
    }
    // 使用后端生成的用户名
    const username = getUsername()
    userInfo.value.name = username || '用户'

    // 加载企业认证状态
    await loadCompanyStatus()
  } catch (error) {
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="w-[80%] mx-auto py-6">
    <div class="flex gap-8">
      <!-- 左侧个人信息 -->
      <div class="w-80 shrink-0">
        <div class="bg-white rounded-lg shadow-sm overflow-hidden">
          <!-- 头像区域 -->
          <div class="bg-gradient-to-r from-blue-500 to-cyan-500 p-8 text-center">
            <div class="w-24 h-24 mx-auto rounded-full bg-white text-blue-500 flex items-center justify-center text-2xl font-bold mb-4">
              {{ userInfo.name ? userInfo.name.slice(0, 2) : '用户' }}
            </div>
            <h2 class="text-xl font-bold text-white mb-1">{{ userInfo.name || '用户' }}</h2>
            <p v-if="userInfo.company" class="text-sm text-white/80">{{ userInfo.company }}</p>
            <p v-else class="text-sm text-white/60">个人账户</p>
            <div v-if="userInfo.verified" class="inline-flex items-center gap-1 mt-3 bg-white/20 text-white text-xs px-3 py-1 rounded-full">
              <CheckCircle :size="14" class="inline-block" />已认证企业
            </div>
            <div v-else class="inline-flex items-center gap-1 mt-3 bg-white/10 text-white/80 text-xs px-3 py-1 rounded-full">
              <User :size="14" class="inline-block" />未认证
            </div>
          </div>

          <!-- 基本信息 -->
          <div class="p-6">
            <div class="flex flex-col gap-4 text-sm">
              <div class="flex items-center gap-3 text-gray-600">
                <Phone :size="16" class="w-5 text-gray-400" />
                {{ userInfo.phone }}
              </div>
              <div class="flex items-center gap-3 text-gray-600">
                <Mail :size="16" class="w-5 text-gray-400" />
                {{ userInfo.email }}
              </div>
              <div class="flex items-center gap-3 text-gray-600">
                <MapPin :size="16" class="w-5 text-gray-400" />
                {{ userInfo.location }}
              </div>
              <div class="flex items-center gap-3 text-gray-600">
                <Calendar :size="16" class="w-5 text-gray-400" />
                加入时间: {{ userInfo.joinTime }}
              </div>
            </div>

            <el-button type="primary" plain class="!w-full !mt-6">
              <Pencil :size="14" class="mr-2" />编辑资料
            </el-button>
          </div>
        </div>
      </div>

      <!-- 右侧内容区 -->
      <div class="flex-1 min-w-0 overflow-hidden">
        <!-- 数据统计 (始终显示) -->
        <div class="grid grid-cols-3 gap-4 mb-6">
          <div class="bg-white rounded-lg p-5 shadow-sm text-center">
            <div class="text-3xl font-bold text-green-500 mb-1">{{ stats.totalBids }}</div>
            <div class="text-sm text-gray-500">投标次数</div>
          </div>
          <div class="bg-white rounded-lg p-5 shadow-sm text-center">
            <div class="text-3xl font-bold text-orange-500 mb-1">{{ stats.successRate }}%</div>
            <div class="text-sm text-gray-500">中标率</div>
          </div>
          <div class="bg-white rounded-lg p-5 shadow-sm text-center">
            <div class="text-3xl font-bold text-purple-500 mb-1">{{ stats.completedOrders }}</div>
            <div class="text-sm text-gray-500">完成订单</div>
          </div>
        </div>

        <!-- 滑动容器 - 固定最小高度确保一致性 -->
        <div class="relative min-h-[400px]">
          <Transition :name="slideDirection === 'left' ? 'slide-left' : 'slide-right'" mode="out-in">
            <!-- 主菜单 -->
            <div v-if="currentModule === 'main'" key="main" class="bg-white rounded-lg shadow-sm p-6 min-h-[360px]">
              <h3 class="text-lg font-bold text-gray-800 mb-5">账户设置</h3>
              <div class="grid grid-cols-2 gap-4">
                <button @click="goToModule('profile')" class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
                  <div class="w-10 h-10 rounded-lg bg-blue-100 text-blue-500 flex items-center justify-center">
                    <UserPen :size="20" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-gray-800">个人信息</div>
                    <div class="text-xs text-gray-500">修改头像、昵称等</div>
                  </div>
                  <ChevronRight :size="16" class="text-gray-300" />
                </button>

                <button @click="goToModule('enterprise')" class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
                  <div class="w-10 h-10 rounded-lg bg-green-100 text-green-500 flex items-center justify-center">
                    <Building2 :size="20" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-gray-800">企业认证</div>
                    <div class="text-xs text-gray-500">提交企业资质认证</div>
                  </div>
                  <ChevronRight :size="16" class="text-gray-300" />
                </button>

                <button @click="goToModule('wallet')" class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
                  <div class="w-10 h-10 rounded-lg bg-orange-100 text-orange-500 flex items-center justify-center">
                    <Wallet :size="20" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-gray-800">账户余额</div>
                    <div class="text-xs text-gray-500">充值、提现、账单</div>
                  </div>
                  <ChevronRight :size="16" class="text-gray-300" />
                </button>

                <button @click="goToModule('security')" class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
                  <div class="w-10 h-10 rounded-lg bg-purple-100 text-purple-500 flex items-center justify-center">
                    <Lock :size="20" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-gray-800">安全设置</div>
                    <div class="text-xs text-gray-500">修改密码、绑定手机</div>
                  </div>
                  <ChevronRight :size="16" class="text-gray-300" />
                </button>

                <button @click="goToModule('notification')" class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50 transition-colors text-left">
                  <div class="w-10 h-10 rounded-lg bg-cyan-100 text-cyan-500 flex items-center justify-center">
                    <Bell :size="20" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-gray-800">消息通知</div>
                    <div class="text-xs text-gray-500">设置通知偏好</div>
                  </div>
                  <ChevronRight :size="16" class="text-gray-300" />
                </button>

                <button @click="handleLogout" class="flex items-center gap-4 p-4 rounded-lg border border-gray-100 hover:border-red-200 hover:bg-red-50 transition-colors text-left">
                  <div class="w-10 h-10 rounded-lg bg-red-100 text-red-500 flex items-center justify-center">
                    <LogOut :size="20" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-gray-800">退出登录</div>
                    <div class="text-xs text-gray-500">安全退出当前账号</div>
                  </div>
                </button>
              </div>
            </div>

            <!-- 企业认证模块 -->
            <div v-else-if="currentModule === 'enterprise'" key="enterprise" class="bg-white rounded-lg shadow-sm p-6 min-h-[360px]">
              <div class="flex items-center gap-3 mb-6">
                <button @click="goBack" class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors">
                  <ArrowLeft :size="16" class="text-gray-600" />
                </button>
                <h3 class="text-lg font-bold text-gray-800">企业认证</h3>
              </div>

              <!-- 已认证状态 -->
              <div v-if="companyStatus.hasVerifiedCompany && companyStatus.company" class="text-center py-8">
                <div class="w-20 h-20 mx-auto bg-green-100 rounded-full flex items-center justify-center mb-4">
                  <CheckCircle :size="40" class="text-green-500" />
                </div>
                <h4 class="text-xl font-bold text-gray-800 mb-2">企业认证已通过</h4>
                <div class="bg-gray-50 rounded-lg p-4 mt-6 text-left">
                  <div class="flex items-center gap-3 mb-3">
                    <span class="text-gray-500 w-24">企业名称:</span>
                    <span class="font-medium text-gray-800">{{ companyStatus.company.name }}</span>
                  </div>
                  <div class="flex items-center gap-3 mb-3">
                    <span class="text-gray-500 w-24">营业执照号:</span>
                    <span class="font-medium text-gray-800">{{ companyStatus.company.businessLicenseNo }}</span>
                  </div>
                  <div class="flex items-center gap-3">
                    <span class="text-gray-500 w-24">认证时间:</span>
                    <span class="font-medium text-gray-800">{{ new Date(companyStatus.company.verifiedAt).toLocaleDateString() }}</span>
                  </div>
                </div>
              </div>

              <!-- 审核中状态 -->
              <div v-else-if="companyStatus.pendingApplication" class="text-center py-8">
                <div class="w-20 h-20 mx-auto bg-yellow-100 rounded-full flex items-center justify-center mb-4">
                  <Clock :size="40" class="text-yellow-500" />
                </div>
                <h4 class="text-xl font-bold text-gray-800 mb-2">认证申请审核中</h4>
                <p class="text-gray-500 mb-6">您的企业认证申请正在审核，请耐心等待</p>
                <div class="bg-gray-50 rounded-lg p-4 text-left">
                  <div class="flex items-center gap-3 mb-3">
                    <span class="text-gray-500 w-24">企业名称:</span>
                    <span class="font-medium text-gray-800">{{ companyStatus.pendingApplication.name }}</span>
                  </div>
                  <div class="flex items-center gap-3 mb-3">
                    <span class="text-gray-500 w-24">营业执照号:</span>
                    <span class="font-medium text-gray-800">{{ companyStatus.pendingApplication.businessLicenseNo }}</span>
                  </div>
                  <div class="flex items-center gap-3">
                    <span class="text-gray-500 w-24">申请时间:</span>
                    <span class="font-medium text-gray-800">{{ new Date(companyStatus.pendingApplication.createdAt).toLocaleDateString() }}</span>
                  </div>
                </div>
              </div>

              <!-- 未认证/被拒绝状态 - 显示表单 -->
              <div v-else class="space-y-4">
                <!-- 被拒绝提示 -->
                <div v-if="companyStatus.latestRejected" class="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
                  <div class="flex items-start gap-3">
                    <AlertCircle :size="16" class="text-red-500 mt-0.5" />
                    <div>
                      <p class="text-red-700 font-medium">上次申请被拒绝</p>
                      <p class="text-red-600 text-sm mt-1">拒绝原因: {{ companyStatus.latestRejected.rejectReason || '未说明' }}</p>
                    </div>
                  </div>
                </div>

                <!-- 步骤1: 上传营业执照 -->
                <div v-if="ocrStep === 'upload'">
                  <label class="block text-sm font-medium text-gray-700 mb-1">上传营业执照</label>
                  <input type="file" ref="fileInputRef" @change="handleFileSelect" accept=".jpg,.jpeg,.png,.pdf" class="hidden" />
                  <div @click="triggerFileSelect" class="border-2 border-dashed border-gray-200 rounded-lg p-8 text-center hover:border-blue-400 transition-colors cursor-pointer">
                    <template v-if="recognizing">
                      <Loader2 :size="30" class="text-blue-500 mb-2 animate-spin" />
                      <p class="text-sm text-blue-600 font-medium">正在识别营业执照...</p>
                      <p class="text-xs text-gray-400 mt-1">请稍候</p>
                    </template>
                    <template v-else>
                      <CloudUpload :size="30" class="text-gray-400 mb-2" />
                      <p class="text-sm text-gray-500">点击上传营业执照图片</p>
                      <p class="text-xs text-gray-400 mt-1">支持 JPG、PNG、PDF 格式，不超过 5MB</p>
                      <p class="text-xs text-blue-500 mt-2">系统将自动识别营业执照信息</p>
                    </template>
                  </div>
                </div>

                <!-- 步骤2: 确认/修改识别结果 -->
                <template v-if="ocrStep === 'recognized'">
                  <div class="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-2">
                    <div class="flex items-center gap-2">
                      <Info :size="16" class="text-blue-500" />
                      <span class="text-sm text-blue-700">以下信息由系统自动识别，请确认并修改</span>
                    </div>
                  </div>

                  <div class="flex items-center gap-3 mb-2">
                    <div class="flex items-center gap-2 text-sm text-gray-500">
                      <FileImage :size="16" class="text-green-500" />
                      <span>{{ selectedFileName }}</span>
                    </div>
                    <el-button type="primary" link size="small" @click="resetUpload">重新上传</el-button>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">企业名称 <span class="text-red-500">*</span></label>
                    <el-input v-model="enterpriseForm.companyName" placeholder="请输入企业全称" />
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">统一社会信用代码 <span class="text-red-500">*</span></label>
                    <el-input v-model="enterpriseForm.creditCode" placeholder="请输入18位信用代码" />
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">法定代表人</label>
                    <el-input v-model="enterpriseForm.legalPerson" placeholder="法定代表人姓名" />
                  </div>

                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">注册资本</label>
                      <el-input v-model="enterpriseForm.registeredCapital" placeholder="注册资本" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">成立日期</label>
                      <el-input v-model="enterpriseForm.establishDate" placeholder="成立日期" />
                    </div>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">经营范围</label>
                    <el-input v-model="enterpriseForm.businessScope" type="textarea" :rows="2" placeholder="经营范围" resize="none" />
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">注册地址</label>
                    <el-input v-model="enterpriseForm.address" placeholder="注册地址" />
                  </div>

                  <el-button type="success" @click="submitEnterprise" :disabled="submitting" :loading="submitting" class="!w-full !py-5">
                    {{ submitting ? '提交中...' : '确认并提交认证' }}
                  </el-button>
                </template>
              </div>
            </div>

            <!-- 个人信息模块 -->
            <div v-else-if="currentModule === 'profile'" key="profile" class="bg-white rounded-lg shadow-sm p-6 min-h-[360px]">
              <div class="flex items-center gap-3 mb-6">
                <button @click="goBack" class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors">
                  <ArrowLeft :size="16" class="text-gray-600" />
                </button>
                <h3 class="text-lg font-bold text-gray-800">个人信息</h3>
              </div>
              <div class="space-y-4">
                <div class="flex items-center gap-4">
                  <div class="w-20 h-20 rounded-full bg-blue-100 text-blue-500 flex items-center justify-center text-xl font-bold">
                    {{ userInfo.name ? userInfo.name.slice(0, 2) : '用户' }}
                  </div>
                  <el-button>
                    <Camera :size="14" class="mr-2 text-gray-400" />更换头像
                  </el-button>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">昵称</label>
                  <el-input v-model="userInfo.name" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">手机号码</label>
                  <el-input :model-value="userInfo.phone" disabled />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
                  <el-input v-model="userInfo.email" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">所在地区</label>
                  <el-input v-model="userInfo.location" />
                </div>
                <el-button type="primary" class="!w-full !py-5">
                  保存修改
                </el-button>
              </div>
            </div>

            <!-- 账户余额模块 -->
            <div v-else-if="currentModule === 'wallet'" key="wallet" class="bg-white rounded-lg shadow-sm p-6 min-h-[360px]">
              <div class="flex items-center gap-3 mb-6">
                <button @click="goBack" class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors">
                  <ArrowLeft :size="16" class="text-gray-600" />
                </button>
                <h3 class="text-lg font-bold text-gray-800">账户余额</h3>
              </div>
              <div class="bg-gradient-to-r from-orange-500 to-amber-500 rounded-xl p-6 text-white mb-6">
                <div class="text-sm opacity-80 mb-1">可用余额</div>
                <div class="text-3xl font-bold mb-4">¥ 12,580.00</div>
                <div class="flex gap-3">
                  <el-button class="!bg-white !text-orange-500 !border-0">充值</el-button>
                  <el-button class="!bg-white/20 !text-white !border-0 hover:!bg-white/30">提现</el-button>
                </div>
              </div>
              <div>
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-medium text-gray-800">最近交易</h4>
                  <el-button type="primary" link size="small">查看全部</el-button>
                </div>
                <div class="space-y-3">
                  <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div class="flex items-center gap-3">
                      <div class="w-10 h-10 rounded-full bg-green-100 text-green-500 flex items-center justify-center"><ArrowDown :size="18" /></div>
                      <div><div class="text-sm font-medium">悬赏收入</div><div class="text-xs text-gray-400">2025-01-15 14:30</div></div>
                    </div>
                    <div class="text-green-500 font-medium">+¥2,500.00</div>
                  </div>
                  <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div class="flex items-center gap-3">
                      <div class="w-10 h-10 rounded-full bg-red-100 text-red-500 flex items-center justify-center"><ArrowUp :size="18" /></div>
                      <div><div class="text-sm font-medium">提现</div><div class="text-xs text-gray-400">2025-01-10 09:15</div></div>
                    </div>
                    <div class="text-red-500 font-medium">-¥5,000.00</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 安全设置模块 -->
            <div v-else-if="currentModule === 'security'" key="security" class="bg-white rounded-lg shadow-sm p-6 min-h-[360px]">
              <div class="flex items-center gap-3 mb-6">
                <button @click="goBack" class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors">
                  <ArrowLeft :size="16" class="text-gray-600" />
                </button>
                <h3 class="text-lg font-bold text-gray-800">安全设置</h3>
              </div>
              <div class="space-y-4">
                <div class="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div class="flex items-center gap-4">
                    <div class="w-10 h-10 rounded-lg bg-purple-100 text-purple-500 flex items-center justify-center"><Key :size="18" /></div>
                    <div><div class="font-medium text-gray-800">登录密码</div><div class="text-xs text-gray-500">定期修改密码可以保护账户安全</div></div>
                  </div>
                  <el-button type="primary" plain size="small">修改</el-button>
                </div>
                <div class="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div class="flex items-center gap-4">
                    <div class="w-10 h-10 rounded-lg bg-green-100 text-green-500 flex items-center justify-center"><Smartphone :size="18" /></div>
                    <div><div class="font-medium text-gray-800">手机绑定</div><div class="text-xs text-gray-500">已绑定 138****8888</div></div>
                  </div>
                  <el-button type="primary" plain size="small">更换</el-button>
                </div>
                <div class="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div class="flex items-center gap-4">
                    <div class="w-10 h-10 rounded-lg bg-blue-100 text-blue-500 flex items-center justify-center"><Mail :size="18" /></div>
                    <div><div class="font-medium text-gray-800">邮箱绑定</div><div class="text-xs text-gray-500">已绑定 wang@example.com</div></div>
                  </div>
                  <el-button type="primary" plain size="small">更换</el-button>
                </div>
                <div class="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div class="flex items-center gap-4">
                    <div class="w-10 h-10 rounded-lg bg-cyan-100 text-cyan-500 flex items-center justify-center"><Shield :size="18" /></div>
                    <div><div class="font-medium text-gray-800">两步验证</div><div class="text-xs text-gray-500">未开启，开启后登录需要验证码</div></div>
                  </div>
                  <el-button type="primary" size="small">开启</el-button>
                </div>
              </div>
            </div>

            <!-- 消息通知模块 -->
            <div v-else-if="currentModule === 'notification'" key="notification" class="bg-white rounded-lg shadow-sm p-6 min-h-[360px]">
              <div class="flex items-center gap-3 mb-6">
                <button @click="goBack" class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors">
                  <ArrowLeft :size="16" class="text-gray-600" />
                </button>
                <h3 class="text-lg font-bold text-gray-800">消息通知</h3>
              </div>
              <div class="space-y-4">
                <div class="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div class="flex items-center gap-4">
                    <div class="w-10 h-10 rounded-lg bg-blue-100 text-blue-500 flex items-center justify-center"><MessageCircle :size="18" /></div>
                    <div><div class="font-medium text-gray-800">系统消息</div><div class="text-xs text-gray-500">平台公告、活动通知等</div></div>
                  </div>
                  <el-switch model-value />
                </div>
                <div class="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div class="flex items-center gap-4">
                    <div class="w-10 h-10 rounded-lg bg-green-100 text-green-500 flex items-center justify-center"><Hammer :size="18" /></div>
                    <div><div class="font-medium text-gray-800">悬赏通知</div><div class="text-xs text-gray-500">投标、中标、悬赏状态变更</div></div>
                  </div>
                  <el-switch model-value />
                </div>
                <div class="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div class="flex items-center gap-4">
                    <div class="w-10 h-10 rounded-lg bg-orange-100 text-orange-500 flex items-center justify-center"><Banknote :size="18" /></div>
                    <div><div class="font-medium text-gray-800">交易通知</div><div class="text-xs text-gray-500">充值、提现、收款提醒</div></div>
                  </div>
                  <el-switch model-value />
                </div>
                <div class="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div class="flex items-center gap-4">
                    <div class="w-10 h-10 rounded-lg bg-purple-100 text-purple-500 flex items-center justify-center"><MessageSquare :size="18" /></div>
                    <div><div class="font-medium text-gray-800">短信通知</div><div class="text-xs text-gray-500">重要消息通过短信提醒</div></div>
                  </div>
                  <el-switch />
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 左滑动画 - 进入子模块 */
.slide-left-enter-active,
.slide-left-leave-active {
  transition: all 0.3s ease-out;
}

.slide-left-enter-from {
  opacity: 0;
  transform: translateX(30px);
}

.slide-left-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}

/* 右滑动画 - 返回主菜单 */
.slide-right-enter-active,
.slide-right-leave-active {
  transition: all 0.3s ease-out;
}

.slide-right-enter-from {
  opacity: 0;
  transform: translateX(-30px);
}

.slide-right-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
