export default {
  teamAdmin: {
    title: '团队管理', description: '审核团队创建与扩容，管理成员上限、资金与准入条件',
    stats: { total: '团队总数', pending: '待处理申请', active: '当前页正常团队', reviewRequired: '当前页待复审' },
    tabs: { teams: '团队', applications: '申请', settings: '准入与升级条件' },
    search: '搜索团队名称或创建者邮箱', allStatuses: '全部状态', active: '正常', frozen: '冻结',
    team: '团队', owner: '创建者', level: '等级', members: '成员', balance: '团队资金', recharge: '有效充值', spend7d: '近 7 天消费',
    reviewRequired: '待复审', empty: '暂无团队', type: '类型', applicant: '申请人', eligibility: '注册 / 充值', days: '天', targetLimit: '目标人数', reason: '理由',
    applicationType: { create: '创建团队', expand: '团队扩容' }, review: '审核', noApplications: '暂无待处理申请',
    settings: { registrationDays: '最低注册天数', minRecharge: '最低累计充值', mode: '条件组合' },
    saveLimit: '保存上限', completeReview: '完成复审', freeze: '冻结团队', restore: '恢复团队', transferable: '可转赠',
    fundLedger: '团队资金流水', reviewReason: '审核原因', waive: '豁免创建门槛（必须填写原因）',
    removeConfirm: '确定移除该成员吗？', operationFailed: '团队管理操作失败', saved: '团队治理配置已保存'
  }
}
