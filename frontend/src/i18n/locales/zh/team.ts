export default {
  team: {
    title: '我的团队',
    description: '创建或加入团队，与成员共享余额和用量查看',
    loadError: '加载团队信息失败',
    role: {
      owner: '发起人',
      member: '成员'
    },
    create: {
      title: '创建团队',
      description: '成为发起人，邀请成员加入',
      namePlaceholder: '团队名称',
      button: '创建团队',
      creating: '创建中...',
      nameRequired: '请输入团队名称',
      reasonPlaceholder: '创建团队的用途或说明',
      additionalInfoPlaceholder: '补充信息（可选）',
	  eligibility: '注册 {days}/{requiredDays} 天，充值 {recharge}/{requiredRecharge}',
      success: '创建申请已提交',
      error: '创建团队失败'
    },
    join: {
      title: '加入团队',
      description: '输入发起人给你的团队码即可加入',
      codePlaceholder: '团队码',
      button: '加入团队',
      joining: '加入中...',
      codeRequired: '请输入团队码',
      messagePlaceholder: '给团队发起人的申请说明（可选）',
      success: '加入申请已提交',
      error: '加入团队失败'
    },
    inviteCode: {
      refresh: '刷新团队码',
      refreshSuccess: '团队码已刷新',
      refreshError: '刷新团队码失败',
      copied: '团队码已复制'
    },
    leave: {
      button: '退出团队',
      leaving: '退出中...',
      confirm: '确定要退出当前团队吗？',
      success: '已退出团队',
      error: '退出团队失败'
    },
    members: {
      title: '团队成员',
      loadError: '加载成员列表失败',
      empty: '暂无团队成员',
      searchPlaceholder: '搜索成员邮箱或用户名',
      balanceHidden: '保密',
      usageHidden: '保密',
      remove: '移除',
      removeConfirm: '确定要移除成员 {email} 吗？',
      removeSuccess: '成员已移除',
      removeError: '移除成员失败',
      count: '共 {count} 位成员',
      columns: {
        user: '用户',
        balance: '余额',
        usage: '累计用量',
        actions: '操作'
      },
      usage: {
        startDate: '开始日期',
        endDate: '结束日期',
        query: '查询消费记录',
        empty: '该时间段内没有消费记录',
        loadError: '加载消费记录失败',
        noPermission: '无权限查看',
        time: '时间',
        model: '模型',
        type: '类型',
        tokens: 'Tokens',
        cost: '费用',
        duration: '耗时'
      }
    },
    fund: {
      balance: '团队资金',
      depositButton: '存入资金',
      depositTitle: '存入团队资金',
      depositHint: '从你的个人余额存入团队资金池',
	  transferable: '当前可转赠：{amount}',
      allocateButton: '分配',
      allocateTitle: '分配团队资金',
      allocateTo: '从团队资金分配给 {email}',
      amount: '金额',
      amountPlaceholder: '请输入金额',
      password: '登录密码确认',
      passwordPlaceholder: '输入你的登录密码',
      confirm: '确认',
      submitting: '处理中...',
      depositSuccess: '已存入团队资金',
      allocateSuccess: '分配成功',
      error: '操作失败',
      amountRequired: '请输入有效金额',
      passwordRequired: '请输入登录密码'
    },
    transfer: {
      button: '划拨余额',
      title: '划拨余额',
      to: '给 {email} 划拨余额',
      amount: '划拨金额',
      amountPlaceholder: '请输入金额',
      password: '登录密码确认',
      passwordPlaceholder: '输入你的登录密码',
      confirm: '确认划拨',
      transferring: '划拨中...',
      amountRequired: '请输入有效的划拨金额',
      passwordRequired: '请输入登录密码',
      success: '余额划拨成功',
      error: '余额划拨失败'
    },
    application: {
      pending: '创建申请待管理员审核',
      approved: '创建申请已批准',
      rejected: '创建申请已拒绝，可补充后重新提交'
    },
    governance: {
      title: '团队等级与扩容',
      level: '团队等级',
      memberLimit: '成员数 / 上限',
      recharge: '有效累计充值',
      spend7d: '近 7 天消费',
      reviewRequired: '现有团队需管理员复审后才能自助扩容',
      upgrade: '检查升级',
      upgradeSuccess: '团队等级已更新',
      upgradeError: '当前不满足更高等级条件',
      targetLimit: '目标人数',
      expandReason: '扩容理由',
      requestExpansion: '申请扩容',
      expandSuccess: '扩容申请已提交',
      expandError: '提交扩容申请失败'
    },
    joinRequests: {
      title: '待处理加入申请',
      reviewError: '处理加入申请失败'
    }
  },
}
