// 悬赏类型中英文对照
export const bountyTypeMap = {
  woven: '梭织',
  knitted: '针织'
}

// 投标状态中英文对照
export const bidStatusMap = {
  pending: '审核中',
  in_progress: '进行中',
  pending_acceptance: '待验收',
  completed: '已完成'
}

export const formatDateTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).replace(/\//g, '-')
}

export const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).replace(/\//g, '-')
}

export const formatComposition = (composition) => {
  if (!composition) return ''
  if (typeof composition === 'string') return composition
  if (typeof composition === 'object') {
    return Object.entries(composition)
      .map(([name, pct]) => {
        const percent = pct * 100
        return `${name} ${percent % 1 === 0 ? percent.toFixed(0) : percent.toFixed(1)}%`
      })
      .join(' / ')
  }
  return ''
}

export const formatMoney = (amount) => {
  return new Intl.NumberFormat('zh-CN').format(amount)
}

export const getStatusText = (status) => {
  return bidStatusMap[status] || status
}
