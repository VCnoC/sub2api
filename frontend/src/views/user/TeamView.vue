<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Page Top Header -->
      <div class="relative overflow-hidden rounded-3xl bg-gradient-to-r from-primary-600 via-primary-500 to-indigo-600 p-6 text-white shadow-xl shadow-primary-500/10 sm:p-8 dark:from-primary-900/90 dark:via-primary-800/80 dark:to-indigo-950/90 dark:shadow-none">
        <div class="absolute -right-10 -top-10 h-64 w-64 rounded-full bg-white/10 blur-3xl"></div>
        <div class="absolute -bottom-10 right-20 h-48 w-48 rounded-full bg-indigo-400/20 blur-2xl"></div>

        <div class="relative z-10 flex flex-col justify-between gap-4 md:flex-row md:items-center">
          <div>
            <div class="inline-flex items-center gap-2 rounded-full bg-white/15 px-3 py-1 text-xs font-medium text-white backdrop-blur-md dark:bg-white/10">
              <Icon name="users" size="xs" />
              <span>{{ t('team.title') }}</span>
            </div>
            <h1 class="mt-2 text-2xl font-extrabold tracking-tight text-white sm:text-3xl">
              {{ team ? team.name : t('team.title') }}
            </h1>
            <p class="mt-1 max-w-2xl text-sm text-primary-100/90 dark:text-primary-200/80">
              {{ t('team.description') }}
            </p>
          </div>

          <!-- Quick status or action badge if in team -->
          <div v-if="team" class="flex flex-wrap items-center gap-3">
            <div class="flex items-center gap-3 rounded-2xl bg-white/10 p-3 backdrop-blur-md border border-white/15 dark:bg-dark-900/40 dark:border-white/10">
              <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-400/20 text-emerald-300 dark:bg-emerald-500/20 dark:text-emerald-400">
                <Icon name="creditCard" size="md" />
              </div>
              <div>
                <div class="text-xs text-primary-100/80 dark:text-gray-400">{{ t('team.fund.balance') }}</div>
                <div class="font-mono text-lg font-bold text-white">{{ formatCurrency(team.balance) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex flex-col items-center justify-center py-20">
        <div class="h-10 w-10 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
        <p class="mt-4 text-sm text-gray-500 dark:text-dark-400 animate-pulse">正在加载团队信息...</p>
      </div>

      <!-- Not in any team -->
      <template v-else-if="!team">
        <!-- Application Status Alert -->
        <div
          v-if="createApplication"
          class="relative overflow-hidden rounded-2xl border p-5 shadow-sm backdrop-blur-xl transition-all duration-300"
          :class="createApplication.status === 'rejected'
            ? 'border-red-200 bg-red-50/80 dark:border-red-900/40 dark:bg-red-950/30'
            : 'border-amber-200 bg-amber-50/80 dark:border-amber-900/40 dark:bg-amber-950/30'"
        >
          <div class="flex flex-wrap items-center justify-between gap-4">
            <div class="flex items-start gap-3">
              <div
                class="mt-0.5 flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl"
                :class="createApplication.status === 'rejected'
                  ? 'bg-red-100 text-red-600 dark:bg-red-900/50 dark:text-red-400'
                  : 'bg-amber-100 text-amber-600 dark:bg-amber-900/50 dark:text-amber-400'"
              >
                <Icon :name="createApplication.status === 'rejected' ? 'exclamationTriangle' : 'clock'" size="md" />
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="font-semibold text-gray-900 dark:text-white">
                    {{ t(`team.application.${createApplication.status}`) }}
                  </h3>
                  <span
                    class="badge"
                    :class="createApplication.status === 'rejected' ? 'badge-danger' : 'badge-warning'"
                  >
                    {{ createApplication.team_name }}
                  </span>
                </div>
                <p v-if="createApplication.review_reason" class="mt-1 text-sm text-gray-600 dark:text-dark-300">
                  {{ createApplication.review_reason }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Create or Join Team Dual Column Cards -->
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <!-- Create Team Card -->
          <div class="card relative overflow-hidden p-6 transition-all duration-300 hover:shadow-lg dark:hover:shadow-primary-900/10">
            <div class="absolute -right-12 -top-12 h-40 w-48 rounded-full bg-primary-500/5 blur-2xl"></div>

            <div class="flex items-center gap-4">
              <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-indigo-600 text-white shadow-md shadow-primary-500/20">
                <Icon name="sparkles" size="md" />
              </div>
              <div>
                <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('team.create.title') }}</h2>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('team.create.description') }}</p>
              </div>
            </div>

            <!-- Creation Eligibility Indicator Box -->
            <div
              v-if="creationEligibility"
              class="mt-5 rounded-2xl border p-4 transition-colors"
              :class="creationEligibility.eligible
                ? 'border-emerald-200 bg-emerald-50/60 dark:border-emerald-900/40 dark:bg-emerald-950/20'
                : 'border-amber-200 bg-amber-50/60 dark:border-amber-900/40 dark:bg-amber-950/20'"
            >
              <div class="flex items-center gap-2 text-xs font-semibold" :class="creationEligibility.eligible ? 'text-emerald-700 dark:text-emerald-400' : 'text-amber-700 dark:text-amber-400'">
                <Icon :name="creationEligibility.eligible ? 'checkCircle' : 'infoCircle'" size="sm" />
                <span>{{ creationEligibility.eligible ? '已满足团队创建条件' : '暂未达到自助创建门槛' }}</span>
              </div>
              <p class="mt-1.5 text-xs leading-relaxed text-gray-600 dark:text-gray-300">
                {{ t('team.create.eligibility', {
                  days: creationEligibility.registration_days,
                  requiredDays: creationEligibility.settings.min_registration_days,
                  recharge: formatCurrency(creationEligibility.effective_recharge),
                  requiredRecharge: formatCurrency(creationEligibility.settings.min_total_recharge)
                }) }}
              </p>
            </div>

            <!-- Form -->
            <div class="mt-5 space-y-4">
              <div>
                <label class="input-label">{{ t('team.create.namePlaceholder') }} <span class="text-red-500">*</span></label>
                <div class="relative">
                  <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-gray-400">
                    <Icon name="users" size="sm" />
                  </div>
                  <input
                    v-model="createName"
                    type="text"
                    class="input w-full pl-10"
                    :placeholder="t('team.create.namePlaceholder')"
                    maxlength="100"
                  />
                </div>
              </div>

              <div>
                <label class="input-label">{{ t('team.create.reasonPlaceholder') }}</label>
                <textarea
                  v-model="createReason"
                  class="input min-h-[80px] w-full resize-y py-2.5"
                  :placeholder="t('team.create.reasonPlaceholder')"
                  maxlength="2000"
                ></textarea>
              </div>

              <div>
                <label class="input-label">{{ t('team.create.additionalInfoPlaceholder') }}</label>
                <textarea
                  v-model="createAdditionalInfo"
                  class="input min-h-[70px] w-full resize-y py-2.5"
                  :placeholder="t('team.create.additionalInfoPlaceholder')"
                  maxlength="4000"
                ></textarea>
              </div>

              <button class="btn btn-primary w-full py-3 shadow-md" :disabled="creating" @click="handleCreate">
                <Icon v-if="creating" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="plus" size="sm" />
                <span>{{ creating ? t('team.create.creating') : t('team.create.button') }}</span>
              </button>
            </div>
          </div>

          <!-- Join Team Card -->
          <div class="card relative overflow-hidden p-6 transition-all duration-300 hover:shadow-lg dark:hover:shadow-emerald-900/10">
            <div class="absolute -right-12 -top-12 h-40 w-48 rounded-full bg-emerald-500/5 blur-2xl"></div>

            <div class="flex items-center gap-4">
              <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 text-white shadow-md shadow-emerald-500/20">
                <Icon name="userPlus" size="md" />
              </div>
              <div>
                <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('team.join.title') }}</h2>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('team.join.description') }}</p>
              </div>
            </div>

            <!-- Hint Box -->
            <div class="mt-5 rounded-2xl border border-gray-200/80 bg-gray-50/60 p-4 dark:border-dark-700/80 dark:bg-dark-800/40">
              <div class="flex items-center gap-2 text-xs font-semibold text-gray-700 dark:text-gray-300">
                <Icon name="key" size="sm" class="text-emerald-500" />
                <span>获取团队邀请码即可快捷申请</span>
              </div>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                向团队发起人索取专属团队码，填入下方输入框提交加入申请，等待发起人确认审核通过。
              </p>
            </div>

            <!-- Form -->
            <div class="mt-5 space-y-4">
              <div>
                <label class="input-label">{{ t('team.join.codePlaceholder') }} <span class="text-red-500">*</span></label>
                <div class="relative">
                  <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-gray-400">
                    <Icon name="key" size="sm" />
                  </div>
                  <input
                    v-model="joinCode"
                    type="text"
                    class="input w-full pl-10 font-mono font-semibold tracking-wide uppercase placeholder:font-sans placeholder:normal-case"
                    :placeholder="t('team.join.codePlaceholder')"
                  />
                </div>
              </div>

              <div>
                <label class="input-label">{{ t('team.join.messagePlaceholder') }}</label>
                <textarea
                  v-model="joinMessage"
                  class="input min-h-[162px] w-full resize-y py-2.5"
                  :placeholder="t('team.join.messagePlaceholder')"
                  maxlength="1000"
                ></textarea>
              </div>

              <button class="btn btn-success w-full py-3 shadow-md" :disabled="joining" @click="handleJoin">
                <Icon v-if="joining" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="arrowRight" size="sm" />
                <span>{{ joining ? t('team.join.joining') : t('team.join.button') }}</span>
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- In a team Workspace -->
      <template v-else>
        <!-- Team Header / Quick Bar Card -->
        <div class="card overflow-hidden p-6 shadow-sm">
          <div class="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
            <!-- Left Info -->
            <div class="flex items-start gap-4">
              <div class="flex h-14 w-14 flex-shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-indigo-600 text-2xl font-black text-white shadow-lg shadow-primary-500/20">
                {{ team.name.charAt(0).toUpperCase() }}
              </div>
              <div>
                <div class="flex flex-wrap items-center gap-2.5">
                  <h2 class="text-xl font-extrabold text-gray-900 dark:text-white">{{ team.name }}</h2>
                  <span
                    class="badge"
                    :class="team.role === 'owner' ? 'badge-primary' : 'badge-gray'"
                  >
                    <Icon :name="team.role === 'owner' ? 'badge' : 'user'" size="xs" />
                    <span>{{ team.role === 'owner' ? t('team.role.owner') : t('team.role.member') }}</span>
                  </span>
                  <span class="badge badge-purple">
                    <Icon name="sparkles" size="xs" />
                    <span>Level {{ team.level }}</span>
                  </span>
                </div>
                <div class="mt-2 flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-dark-400">
                  <span class="inline-flex items-center gap-1">
                    <Icon name="users" size="xs" />
                    {{ t('team.members.count', { count: filteredMembers.length }) }} / {{ team.member_limit }}
                  </span>
                  <span v-if="team.created_at" class="inline-flex items-center gap-1">
                    <Icon name="calendar" size="xs" />
                    创建于 {{ formatDateTime(team.created_at) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Right Actions & Fund Pool -->
            <div class="flex flex-wrap items-center gap-3">
              <!-- Fund Pool Indicator -->
              <div class="flex items-center gap-3 rounded-2xl border border-emerald-200 bg-emerald-50/80 px-4 py-2.5 dark:border-emerald-800/60 dark:bg-emerald-950/40">
                <div>
                  <div class="text-xs text-emerald-800/80 dark:text-emerald-300/80">{{ t('team.fund.balance') }}</div>
                  <div class="font-mono text-base font-extrabold text-emerald-600 dark:text-emerald-400">
                    {{ formatCurrency(team.balance) }}
                  </div>
                </div>
                <button class="btn btn-success btn-sm ml-1" @click="openDeposit">
                  <Icon name="creditCard" size="xs" />
                  <span>{{ t('team.fund.depositButton') }}</span>
                </button>
              </div>

              <!-- Owner Invite Code Controls -->
              <template v-if="isOwner">
                <div class="flex items-center gap-2 rounded-2xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-800">
                  <span class="text-xs text-gray-400 dark:text-dark-400">团队码:</span>
                  <code class="font-mono text-sm font-bold text-gray-900 dark:text-white">{{ team.invite_code }}</code>
                  <button class="btn btn-secondary btn-sm p-1.5" title="复制团队码" @click="copyCode">
                    <Icon name="copy" size="xs" />
                  </button>
                </div>

                <button class="btn btn-secondary" :disabled="refreshing" title="重置团队邀请码" @click="handleRefreshCode">
                  <Icon v-if="refreshing" name="refresh" size="sm" class="animate-spin" />
                  <Icon v-else name="refresh" size="sm" />
                  <span class="hidden sm:inline">{{ t('team.inviteCode.refresh') }}</span>
                </button>
              </template>

              <!-- Leave Team Button for Member -->
              <button
                v-else
                class="btn btn-danger-outline"
                :disabled="leaving"
                @click="handleLeave"
              >
                <Icon v-if="leaving" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="xCircle" size="sm" />
                <span>{{ leaving ? t('team.leave.leaving') : t('team.leave.button') }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- 4 Stats Cards Grid -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <!-- Card 1: Level -->
          <div class="stat-card transition-all duration-300 hover:-translate-y-0.5">
            <div class="stat-icon bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400">
              <Icon name="badge" size="md" />
            </div>
            <div>
              <div class="stat-label">{{ t('team.governance.level') }}</div>
              <div class="stat-value text-purple-600 dark:text-purple-400">Lv. {{ team.level }}</div>
              <div class="mt-1 text-xs text-gray-400">资源组共享等级</div>
            </div>
          </div>

          <!-- Card 2: Members -->
          <div class="stat-card transition-all duration-300 hover:-translate-y-0.5">
            <div class="stat-icon bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
              <Icon name="users" size="md" />
            </div>
            <div class="w-full min-w-0">
              <div class="stat-label">{{ t('team.governance.memberLimit') }}</div>
              <div class="stat-value text-gray-900 dark:text-white">
                {{ team.member_count }} <span class="text-sm font-normal text-gray-400">/ {{ team.member_limit }}</span>
              </div>
              <div class="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-dark-800">
                <div
                  class="h-full rounded-full bg-blue-500 transition-all duration-500"
                  :style="{ width: Math.min(100, Math.round((team.member_count / (team.member_limit || 1)) * 100)) + '%' }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Card 3: Recharge -->
          <div class="stat-card transition-all duration-300 hover:-translate-y-0.5">
            <div class="stat-icon bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400">
              <Icon name="trendingUp" size="md" />
            </div>
            <div>
              <div class="stat-label">{{ t('team.governance.recharge') }}</div>
              <div class="stat-value text-emerald-600 dark:text-emerald-400">
                {{ formatCurrency(team.effective_recharge) }}
              </div>
              <div class="mt-1 text-xs text-gray-400">团队累计有效充值</div>
            </div>
          </div>

          <!-- Card 4: Spend 7d -->
          <div class="stat-card transition-all duration-300 hover:-translate-y-0.5">
            <div class="stat-icon bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400">
              <Icon name="chartBar" size="md" />
            </div>
            <div>
              <div class="stat-label">{{ t('team.governance.spend7d') }}</div>
              <div class="stat-value text-amber-600 dark:text-amber-400">
                {{ formatCurrency(team.spend_7d) }}
              </div>
              <div class="mt-1 text-xs text-gray-400">近 7 天总 API 消费</div>
            </div>
          </div>
        </div>

        <!-- Governance & Expansion Box (Owner only) -->
        <div v-if="isOwner" class="card p-6 shadow-sm">
          <div class="flex flex-wrap items-center justify-between gap-4 border-b border-gray-100 pb-4 dark:border-dark-800">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400">
                <Icon name="shield" size="md" />
              </div>
              <div>
                <h3 class="font-bold text-gray-900 dark:text-white">{{ t('team.governance.title') }}</h3>
                <p v-if="team.review_required" class="mt-0.5 text-xs text-amber-600 dark:text-amber-400">
                  <Icon name="exclamationTriangle" size="xs" class="inline mr-1" />
                  {{ t('team.governance.reviewRequired') }}
                </p>
                <p v-else class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">按需申请提高团队上限或升级团队共享等级</p>
              </div>
            </div>

            <button
              class="btn btn-secondary"
              :disabled="upgrading || team.review_required"
              @click="handleUpgrade"
            >
              <Icon v-if="upgrading" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="arrowUp" size="sm" class="text-purple-500" />
              <span>{{ t('team.governance.upgrade') }}</span>
            </button>
          </div>

          <div class="mt-4 grid gap-3 md:grid-cols-[180px_1fr_auto]">
            <div>
              <label class="input-label text-xs">{{ t('team.governance.targetLimit') }}</label>
              <input
                v-model.number="expandForm.target_limit"
                type="number"
                min="41"
                class="input"
                :placeholder="t('team.governance.targetLimit')"
              />
            </div>
            <div>
              <label class="input-label text-xs">{{ t('team.governance.expandReason') }}</label>
              <input
                v-model="expandForm.reason"
                class="input"
                :placeholder="t('team.governance.expandReason')"
              />
            </div>
            <div class="flex items-end">
              <button
                class="btn btn-primary w-full md:w-auto"
                :disabled="expanding || team.review_required"
                @click="handleExpansion"
              >
                <Icon v-if="expanding" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="plus" size="sm" />
                <span>{{ t('team.governance.requestExpansion') }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Join Requests List (Owner only) -->
        <div v-if="isOwner && joinRequests.length" class="card p-6 border-l-4 border-l-primary-500 shadow-sm">
          <div class="flex items-center gap-2">
            <div class="relative flex h-3 w-3">
              <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary-400 opacity-75"></span>
              <span class="relative inline-flex h-3 w-3 rounded-full bg-primary-500"></span>
            </div>
            <h3 class="font-bold text-gray-900 dark:text-white">{{ t('team.joinRequests.title') }}</h3>
            <span class="badge badge-primary ml-1">{{ joinRequests.length }}</span>
          </div>

          <div class="mt-4 divide-y divide-gray-100 dark:divide-dark-800">
            <div
              v-for="request in joinRequests"
              :key="request.id"
              class="flex flex-wrap items-center justify-between gap-4 py-3.5 transition-colors hover:bg-gray-50/50 dark:hover:bg-dark-800/30"
            >
              <div class="flex items-center gap-3">
                <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-gray-100 font-bold text-gray-600 dark:bg-dark-800 dark:text-dark-300">
                  {{ request.applicant_email.charAt(0).toUpperCase() }}
                </div>
                <div>
                  <p class="font-semibold text-gray-900 dark:text-white">{{ request.applicant_email }}</p>
                  <p v-if="request.message" class="text-xs text-gray-500 dark:text-dark-400">申请说明: {{ request.message }}</p>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <button class="btn btn-secondary btn-sm" @click="handleJoinReview(request.id, false)">
                  <Icon name="x" size="xs" />
                  <span>{{ t('common.reject') }}</span>
                </button>
                <button class="btn btn-primary btn-sm" @click="handleJoinReview(request.id, true)">
                  <Icon name="check" size="xs" />
                  <span>{{ t('common.approve') }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Member List Table Card -->
        <div class="card p-6 shadow-sm">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-primary-100 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400">
                <Icon name="users" size="md" />
              </div>
              <div>
                <h3 class="text-base font-bold text-gray-900 dark:text-white">{{ t('team.members.title') }}</h3>
                <p class="text-xs text-gray-500 dark:text-dark-400">查看团队成员、管理分配额度及详细消费日志</p>
              </div>
            </div>

            <!-- Member Search Input -->
            <div class="relative w-full sm:w-64">
              <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-gray-400">
                <Icon name="search" size="sm" />
              </div>
              <input
                v-model="memberSearch"
                type="text"
                class="input w-full pl-9"
                :placeholder="t('team.members.searchPlaceholder')"
              />
            </div>
          </div>

          <!-- Members Table Loading -->
          <div v-if="membersLoading" class="flex justify-center py-12">
            <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>

          <!-- Empty Members -->
          <div
            v-else-if="filteredMembers.length === 0"
            class="mt-6 rounded-2xl border border-dashed border-gray-200 p-8 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400"
          >
            <Icon name="users" size="xl" class="mx-auto mb-2 text-gray-300 dark:text-dark-600" />
            <p>{{ t('team.members.empty') }}</p>
          </div>

          <!-- Members Table -->
          <div v-else class="mt-6 overflow-x-auto rounded-2xl border border-gray-100 dark:border-dark-800">
            <table class="w-full min-w-[680px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-100 bg-gray-50/80 text-xs font-semibold text-gray-500 dark:border-dark-800 dark:bg-dark-800/50 dark:text-dark-400">
                  <th class="px-4 py-3.5">{{ t('team.members.columns.user') }}</th>
                  <th class="px-4 py-3.5 text-right">{{ t('team.members.columns.balance') }}</th>
                  <th class="px-4 py-3.5 text-right">{{ t('team.members.columns.usage') }}</th>
                  <th v-if="isOwner" class="px-4 py-3.5 text-right">{{ t('team.members.columns.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
                <template v-for="member in filteredMembers" :key="member.id">
                  <tr
                    class="group cursor-pointer transition-colors hover:bg-gray-50/80 dark:hover:bg-dark-800/40"
                    :class="expandedMemberId === member.id ? 'bg-primary-50/30 dark:bg-primary-950/10' : ''"
                    @click="toggleMember(member)"
                  >
                    <!-- User Email & Role -->
                    <td class="px-4 py-4">
                      <div class="flex items-center gap-3">
                        <div
                          class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg text-gray-400 transition-transform duration-200 group-hover:text-primary-500"
                          :class="expandedMemberId === member.id ? 'rotate-90 text-primary-500' : ''"
                        >
                          <Icon name="chevronRight" size="sm" />
                        </div>
                        <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
                          {{ member.email.charAt(0).toUpperCase() }}
                        </div>
                        <div>
                          <div class="font-semibold text-gray-900 dark:text-white flex items-center gap-2">
                            <span>{{ member.email }}</span>
                            <span v-if="member.id === currentUserId" class="text-xs font-normal text-primary-500">(我自己)</span>
                          </div>
                          <span
                            class="badge mt-1"
                            :class="member.role === 'owner' ? 'badge-primary' : 'badge-gray'"
                          >
                            {{ member.role === 'owner' ? t('team.role.owner') : t('team.role.member') }}
                          </span>
                        </div>
                      </div>
                    </td>

                    <!-- Balance -->
                    <td class="px-4 py-4 text-right font-mono text-sm font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ member.balance === null ? t('team.members.balanceHidden') : formatCurrency(member.balance) }}
                    </td>

                    <!-- Total Usage -->
                    <td class="px-4 py-4 text-right font-mono text-sm tabular-nums text-gray-600 dark:text-gray-300">
                      {{ member.total_usage === null ? t('team.members.usageHidden') : formatCurrency(member.total_usage) }}
                    </td>

                    <!-- Owner Actions -->
                    <td v-if="isOwner" class="px-4 py-4 text-right">
                      <div class="flex items-center justify-end gap-2" @click.stop>
                        <button class="btn btn-secondary btn-sm" @click="openAllocate(member)">
                          <Icon name="creditCard" size="xs" />
                          <span>{{ t('team.fund.allocateButton') }}</span>
                        </button>
                        <template v-if="member.id !== currentUserId">
                          <button class="btn btn-secondary btn-sm" @click="openTransfer(member)">
                            <Icon name="swap" size="xs" />
                            <span>{{ t('team.transfer.button') }}</span>
                          </button>
                          <button
                            class="btn btn-danger-outline btn-sm"
                            :disabled="removingId === member.id"
                            @click="handleRemove(member)"
                          >
                            <Icon v-if="removingId === member.id" name="refresh" size="xs" class="animate-spin" />
                            <Icon v-else name="trash" size="xs" />
                            <span>{{ t('team.members.remove') }}</span>
                          </button>
                        </template>
                      </div>
                    </td>
                  </tr>

                  <!-- Expanded Usage Details Panel -->
                  <tr v-if="expandedMemberId === member.id" class="bg-gray-50/70 dark:bg-dark-900/60">
                    <td :colspan="isOwner ? 4 : 3" class="p-4 sm:p-6">
                      <div v-if="!canViewUsage(member)" class="rounded-2xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
                        <Icon name="lock" size="md" class="mx-auto mb-2 text-gray-400" />
                        {{ t('team.members.usage.noPermission') }}
                      </div>

                      <div v-else class="space-y-4">
                        <!-- Date Selector Bar -->
                        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                          <div class="flex flex-wrap gap-1.5">
                            <button
                              v-for="preset in datePresets"
                              :key="preset.value"
                              class="btn btn-sm"
                              :class="activePreset === preset.value ? 'btn-primary' : 'btn-secondary'"
                              @click="applyDatePreset(preset.value)"
                            >
                              {{ t(preset.labelKey) }}
                            </button>
                          </div>

                          <div class="flex flex-wrap items-end gap-2">
                            <div>
                              <label class="input-label text-xs mb-1">{{ t('team.members.usage.startDate') }}</label>
                              <input v-model="usageStartDate" type="date" class="input py-1.5 text-xs" />
                            </div>
                            <div>
                              <label class="input-label text-xs mb-1">{{ t('team.members.usage.endDate') }}</label>
                              <input v-model="usageEndDate" type="date" class="input py-1.5 text-xs" />
                            </div>
                            <button
                              class="btn btn-secondary btn-sm h-[38px]"
                              :disabled="usageLoading[member.id]"
                              @click="loadMemberUsage(member)"
                            >
                              <Icon v-if="usageLoading[member.id]" name="refresh" size="xs" class="animate-spin" />
                              <Icon v-else name="search" size="xs" />
                              <span>{{ t('team.members.usage.query') }}</span>
                            </button>
                          </div>
                        </div>

                        <!-- Usage Table Loading -->
                        <div v-if="usageLoading[member.id]" class="flex justify-center py-8">
                          <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
                        </div>

                        <!-- Empty Logs -->
                        <div v-else-if="memberUsage[member.id]?.length === 0" class="rounded-xl border border-dashed border-gray-200 p-6 text-center text-xs text-gray-500 dark:border-dark-700 dark:text-dark-400">
                          {{ t('team.members.usage.empty') }}
                        </div>

                        <!-- Usage Logs Table -->
                        <div v-else-if="memberUsage[member.id]?.length > 0" class="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
                          <table class="w-full min-w-[900px] text-left text-xs">
                            <thead>
                              <tr class="border-b border-gray-200 bg-gray-100/80 font-medium text-gray-700 dark:border-dark-700 dark:bg-dark-700/60 dark:text-gray-300">
                                <th class="px-3 py-2.5">{{ t('team.members.usage.time') }}</th>
                                <th class="px-3 py-2.5">{{ t('team.members.usage.model') }}</th>
                                <th class="px-3 py-2.5">{{ t('team.members.usage.type') }}</th>
                                <th class="px-3 py-2.5">{{ t('team.members.usage.tokens') }}</th>
                                <th class="px-3 py-2.5">{{ t('team.members.usage.cost') }}</th>
                                <th class="px-3 py-2.5">首字耗时</th>
                                <th class="px-3 py-2.5">{{ t('team.members.usage.duration') }}</th>
                              </tr>
                            </thead>
                            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
                              <tr
                                v-for="log in memberUsage[member.id]"
                                :key="log.id"
                                class="transition-colors hover:bg-gray-50/50 dark:hover:bg-dark-700/30"
                              >
                                <td class="px-3 py-2.5 text-gray-500 dark:text-gray-400 font-mono">{{ formatDateTime(log.created_at) }}</td>
                                <td class="px-3 py-2.5 font-semibold text-gray-900 dark:text-white">{{ log.model }}</td>
                                <td class="px-3 py-2.5">
                                  <span
                                    class="inline-flex items-center rounded-md px-2 py-0.5 text-[11px] font-medium"
                                    :class="getRequestTypeBadgeClass(log)"
                                  >
                                    {{ getRequestTypeLabel(log) }}
                                  </span>
                                </td>
                                <td class="px-3 py-2.5">
                                  <div class="space-y-0.5 font-mono">
                                    <div class="flex items-center gap-2">
                                      <span class="inline-flex items-center gap-0.5 text-emerald-600 dark:text-emerald-400">
                                        <Icon name="arrowDown" size="xs" />
                                        <span class="tabular-nums">{{ log.input_tokens.toLocaleString() }}</span>
                                      </span>
                                      <span class="inline-flex items-center gap-0.5 text-violet-600 dark:text-violet-400">
                                        <Icon name="arrowUp" size="xs" />
                                        <span class="tabular-nums">{{ log.output_tokens.toLocaleString() }}</span>
                                      </span>
                                    </div>
                                    <div v-if="log.cache_read_tokens > 0 || log.cache_creation_tokens > 0" class="flex items-center gap-2">
                                      <span v-if="log.cache_read_tokens > 0" class="inline-flex items-center gap-0.5 text-sky-600 dark:text-sky-400">
                                        <Icon name="inbox" size="xs" />
                                        <span class="tabular-nums">{{ formatCacheTokens(log.cache_read_tokens) }}</span>
                                      </span>
                                      <span v-if="log.cache_creation_tokens > 0" class="inline-flex items-center gap-0.5 text-amber-600 dark:text-amber-400">
                                        <Icon name="edit" size="xs" />
                                        <span class="tabular-nums">{{ formatCacheTokens(log.cache_creation_tokens) }}</span>
                                      </span>
                                    </div>
                                  </div>
                                </td>
                                <td class="px-3 py-2.5 font-mono font-bold text-emerald-600 dark:text-emerald-400">
                                  ${{ (log.actual_cost ?? 0).toFixed(6) }}
                                </td>
                                <td class="px-3 py-2.5 font-mono text-gray-500">{{ log.first_token_ms != null ? formatDuration(log.first_token_ms) : '-' }}</td>
                                <td class="px-3 py-2.5 font-mono text-gray-500">{{ formatDuration(log.duration_ms) }}</td>
                              </tr>
                            </tbody>
                          </table>
                        </div>

                        <!-- Pagination -->
                        <div class="pt-2">
                          <Pagination
                            v-if="usagePagination[member.id]?.total > 0"
                            :page="usagePagination[member.id].page"
                            :total="usagePagination[member.id].total"
                            :page-size="usagePagination[member.id].page_size"
                            @update:page="(page: number) => loadMemberUsage(member, page)"
                          />
                        </div>
                      </div>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>

    <!-- Modal 1: Balance Transfer Modal -->
    <Teleport to="body">
      <div
        v-if="transferTarget"
        class="modal-overlay"
        @click.self="closeTransfer"
      >
        <div class="modal-content max-w-md">
          <div class="modal-header">
            <div class="flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-100 text-primary-600 dark:bg-primary-900/40 dark:text-primary-400">
                <Icon name="swap" size="sm" />
              </div>
              <h3 class="modal-title">{{ t('team.transfer.title') }}</h3>
            </div>
            <button class="btn btn-ghost btn-icon" @click="closeTransfer">
              <Icon name="x" size="sm" />
            </button>
          </div>

          <div class="modal-body space-y-4">
            <div class="rounded-xl border border-gray-200/80 bg-gray-50/80 p-3.5 dark:border-dark-700/80 dark:bg-dark-900/50">
              <p class="text-xs text-gray-500 dark:text-dark-400">接收用户:</p>
              <p class="font-semibold text-gray-900 dark:text-white">{{ transferTarget.email }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('team.transfer.amount') }} <span class="text-red-500">*</span></label>
              <div class="relative">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-gray-400 font-bold">$</div>
                <input
                  v-model.number="transferAmount"
                  type="number"
                  step="0.01"
                  min="0.01"
                  class="input w-full pl-8 font-mono"
                  :placeholder="t('team.transfer.amountPlaceholder')"
                />
              </div>
            </div>

            <div>
              <label class="input-label">{{ t('team.transfer.password') }} <span class="text-red-500">*</span></label>
              <div class="relative">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-gray-400">
                  <Icon name="lock" size="sm" />
                </div>
                <input
                  v-model="transferPassword"
                  type="password"
                  class="input w-full pl-9"
                  :placeholder="t('team.transfer.passwordPlaceholder')"
                />
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" @click="closeTransfer">{{ t('common.cancel') }}</button>
            <button class="btn btn-primary" :disabled="transferring" @click="handleTransfer">
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <span>{{ transferring ? t('team.transfer.transferring') : t('team.transfer.confirm') }}</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Modal 2: Team Fund Deposit/Allocate Modal -->
    <Teleport to="body">
      <div
        v-if="fundModalOpen"
        class="modal-overlay"
        @click.self="closeFund"
      >
        <div class="modal-content max-w-md">
          <div class="modal-header">
            <div class="flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-emerald-100 text-emerald-600 dark:bg-emerald-900/40 dark:text-emerald-400">
                <Icon name="creditCard" size="sm" />
              </div>
              <h3 class="modal-title">
                {{ fundMode === 'deposit' ? t('team.fund.depositTitle') : t('team.fund.allocateTitle') }}
              </h3>
            </div>
            <button class="btn btn-ghost btn-icon" @click="closeFund">
              <Icon name="x" size="sm" />
            </button>
          </div>

          <div class="modal-body space-y-4">
            <div class="rounded-xl border border-emerald-200/80 bg-emerald-50/80 p-3.5 dark:border-emerald-800/60 dark:bg-emerald-950/40">
              <p class="text-xs text-gray-600 dark:text-dark-300">
                <template v-if="fundMode === 'allocate' && fundTarget">
                  {{ t('team.fund.allocateTo', { email: fundTarget.email }) }}
                </template>
                <template v-else>
                  {{ t('team.fund.depositHint') }}
                </template>
              </p>
              <p v-if="team && fundMode === 'deposit'" class="mt-1 font-mono text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                {{ t('team.fund.transferable', { amount: formatCurrency(team.transferable_balance) }) }}
              </p>
            </div>

            <div>
              <label class="input-label">{{ t('team.fund.amount') }} <span class="text-red-500">*</span></label>
              <div class="relative">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-gray-400 font-bold">$</div>
                <input
                  v-model.number="fundAmount"
                  type="number"
                  step="0.01"
                  min="0.01"
                  class="input w-full pl-8 font-mono"
                  :placeholder="t('team.fund.amountPlaceholder')"
                />
              </div>
            </div>

            <div>
              <label class="input-label">{{ t('team.fund.password') }} <span class="text-red-500">*</span></label>
              <div class="relative">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-gray-400">
                  <Icon name="lock" size="sm" />
                </div>
                <input
                  v-model="fundPassword"
                  type="password"
                  class="input w-full pl-9"
                  :placeholder="t('team.fund.passwordPlaceholder')"
                />
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" @click="closeFund">{{ t('common.cancel') }}</button>
            <button class="btn btn-success" :disabled="fundSubmitting" @click="handleFundSubmit">
              <Icon v-if="fundSubmitting" name="refresh" size="sm" class="animate-spin" />
              <span>{{ fundSubmitting ? t('team.fund.submitting') : t('team.fund.confirm') }}</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { formatCacheTokens } from '@/utils/formatters'
import { extractApiErrorMessage } from '@/utils/apiError'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import {
  createTeam,
  getMyCreateApplication,
  getCreationEligibility,
  getMyTeam,
  joinTeam,
  listJoinRequests,
  reviewJoinRequest,
  upgradeTeam,
  requestTeamExpansion,
  leaveTeam,
  listMembers,
  refreshInviteCode,
  removeMember,
  transferBalance,
  getMemberUsage,
  depositFund,
  allocateFund
} from '@/api/team'
import type { Team, TeamApplication, TeamCreationEligibility, TeamJoinRequest, TeamMember, UsageLog } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const team = ref<Team | null>(null)
const members = ref<TeamMember[]>([])
const membersLoading = ref(false)
const memberSearch = ref('')

const createName = ref('')
const createReason = ref('')
const createAdditionalInfo = ref('')
const createApplication = ref<TeamApplication | null>(null)
const creationEligibility = ref<TeamCreationEligibility | null>(null)
const joinCode = ref('')
const joinMessage = ref('')
const creating = ref(false)
const joining = ref(false)
const leaving = ref(false)
const refreshing = ref(false)
const upgrading = ref(false)
const expanding = ref(false)
const expandForm = ref({ target_limit: 41, reason: '' })
const joinRequests = ref<TeamJoinRequest[]>([])
const removingId = ref<number | null>(null)

const transferTarget = ref<TeamMember | null>(null)
const transferAmount = ref<number>(0)
const transferPassword = ref('')
const transferring = ref(false)

// 团队资金弹窗状态（deposit=成员存入 / allocate=owner 分配）
const fundModalOpen = ref(false)
const fundMode = ref<'deposit' | 'allocate'>('deposit')
const fundTarget = ref<TeamMember | null>(null)
const fundAmount = ref<number>(0)
const fundPassword = ref('')
const fundSubmitting = ref(false)

const expandedMemberId = ref<number | null>(null)
const memberUsage = ref<Record<number, UsageLog[]>>({})
const usageLoading = ref<Record<number, boolean>>({})
const usagePagination = ref<Record<number, { page: number; page_size: number; total: number }>>({})

const now = new Date()
const weekAgo = new Date(now)
weekAgo.setDate(weekAgo.getDate() - 6)
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}
const usageStartDate = ref(formatLocalDate(weekAgo))
const usageEndDate = ref(formatLocalDate(now))
const activePreset = ref('last7Days')

const datePresets = [
  { labelKey: 'dates.today', value: 'today' },
  { labelKey: 'dates.yesterday', value: 'yesterday' },
  { labelKey: 'dates.last7Days', value: 'last7Days' },
  { labelKey: 'dates.last30Days', value: 'last30Days' }
]

function applyDatePreset(value: string) {
  activePreset.value = value
  const today = new Date()
  const start = new Date(today)
  switch (value) {
    case 'today':
      break
    case 'yesterday':
      start.setDate(start.getDate() - 1)
      today.setDate(today.getDate() - 1)
      break
    case 'last7Days':
      start.setDate(start.getDate() - 6)
      break
    case 'last30Days':
      start.setDate(start.getDate() - 29)
      break
  }
  usageStartDate.value = formatLocalDate(start)
  usageEndDate.value = formatLocalDate(today)
}

const currentUserId = computed(() => authStore.user?.id ?? 0)
const isOwner = computed(() => team.value?.role === 'owner')

const filteredMembers = computed(() => {
  const query = memberSearch.value.trim().toLowerCase()
  if (!query) return members.value
  return members.value.filter(
    (m) =>
      m.email.toLowerCase().includes(query) ||
      (m.username && m.username.toLowerCase().includes(query))
  )
})

onMounted(() => {
  loadTeam()
})

function formatDuration(ms: number | null | undefined): string {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms.toFixed(0)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

async function loadTeam() {
  loading.value = true
  try {
    team.value = await getMyTeam()
    if (team.value) {
      await loadMembers()
      if (team.value.role === 'owner') joinRequests.value = await listJoinRequests()
    } else {
      const [application, eligibility] = await Promise.all([getMyCreateApplication(), getCreationEligibility()])
      createApplication.value = application
      creationEligibility.value = eligibility
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.loadError')))
  } finally {
    loading.value = false
  }
}

async function loadMembers() {
  membersLoading.value = true
  try {
    const result = await listMembers({ page: 1, page_size: 100 })
    members.value = result.items
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.members.loadError')))
  } finally {
    membersLoading.value = false
  }
}

// Owner can view any member's usage; regular members can only view their own.
function canViewUsage(member: TeamMember): boolean {
  return isOwner.value || member.id === currentUserId.value
}

function toggleMember(member: TeamMember) {
  if (expandedMemberId.value === member.id) {
    expandedMemberId.value = null
  } else {
    expandedMemberId.value = member.id
    if (canViewUsage(member) && !memberUsage.value[member.id]) {
      loadMemberUsage(member)
    }
  }
}

async function loadMemberUsage(member: TeamMember, page = 1) {
  usageLoading.value[member.id] = true
  try {
    const result = await getMemberUsage(member.id, {
      page,
      page_size: 10,
      start_date: usageStartDate.value,
      end_date: usageEndDate.value
    })
    memberUsage.value[member.id] = result.items
    usagePagination.value[member.id] = {
      page: result.page,
      page_size: result.page_size,
      total: result.total
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.members.usage.loadError')))
  } finally {
    usageLoading.value[member.id] = false
  }
}

function getRequestTypeLabel(log: UsageLog): string {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return t('usage.cyber')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

function getRequestTypeBadgeClass(log: UsageLog): string {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return 'bg-red-100 text-red-800 dark:bg-red-900/60 dark:text-red-200'
  if (requestType === 'ws_v2') return 'bg-violet-100 text-violet-800 dark:bg-violet-900/60 dark:text-violet-200'
  if (requestType === 'stream') return 'bg-blue-100 text-blue-800 dark:bg-blue-900/60 dark:text-blue-200'
  if (requestType === 'sync') return 'bg-gray-100 text-gray-800 dark:bg-gray-700/60 dark:text-gray-200'
  return 'bg-amber-100 text-amber-800 dark:bg-amber-900/60 dark:text-amber-200'
}

async function handleCreate() {
  const name = createName.value.trim()
  if (!name) {
    appStore.showError(t('team.create.nameRequired'))
    return
  }
  creating.value = true
  try {
    createApplication.value = await createTeam(name, createReason.value, createAdditionalInfo.value)
    createName.value = ''
    createReason.value = ''
    createAdditionalInfo.value = ''
    appStore.showSuccess(t('team.create.success'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.create.error')))
  } finally {
    creating.value = false
  }
}

async function handleJoin() {
  const code = joinCode.value.trim()
  if (!code) {
    appStore.showError(t('team.join.codeRequired'))
    return
  }
  joining.value = true
  try {
    await joinTeam(code, joinMessage.value)
    joinCode.value = ''
    joinMessage.value = ''
    appStore.showSuccess(t('team.join.success'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.join.error')))
  } finally {
    joining.value = false
  }
}

async function handleUpgrade() {
  upgrading.value = true
  try {
    await upgradeTeam()
    appStore.showSuccess(t('team.governance.upgradeSuccess'))
    await loadTeam()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.governance.upgradeError')))
  } finally {
    upgrading.value = false
  }
}

async function handleExpansion() {
  if (expandForm.value.target_limit <= 40 || !expandForm.value.reason.trim()) return
  expanding.value = true
  try {
    await requestTeamExpansion(expandForm.value.target_limit, expandForm.value.reason.trim())
    expandForm.value = { target_limit: 41, reason: '' }
    appStore.showSuccess(t('team.governance.expandSuccess'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.governance.expandError')))
  } finally {
    expanding.value = false
  }
}

async function handleJoinReview(id: number, approve: boolean) {
  try {
    await reviewJoinRequest(id, approve)
    joinRequests.value = await listJoinRequests()
    await loadMembers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.joinRequests.reviewError')))
  }
}

async function handleRefreshCode() {
  refreshing.value = true
  try {
    const result = await refreshInviteCode()
    if (team.value) {
      team.value.invite_code = result.invite_code
    }
    appStore.showSuccess(t('team.inviteCode.refreshSuccess'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.inviteCode.refreshError')))
  } finally {
    refreshing.value = false
  }
}

async function handleLeave() {
  if (!confirm(t('team.leave.confirm'))) return
  leaving.value = true
  try {
    await leaveTeam()
    team.value = null
    members.value = []
    appStore.showSuccess(t('team.leave.success'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.leave.error')))
  } finally {
    leaving.value = false
  }
}

async function handleRemove(member: TeamMember) {
  if (!confirm(t('team.members.removeConfirm', { email: member.email }))) return
  removingId.value = member.id
  try {
    await removeMember(member.id)
    appStore.showSuccess(t('team.members.removeSuccess'))
    await loadMembers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.members.removeError')))
  } finally {
    removingId.value = null
  }
}

function openTransfer(member: TeamMember) {
  transferTarget.value = member
  transferAmount.value = 0
  transferPassword.value = ''
}

function closeTransfer() {
  transferTarget.value = null
  transferAmount.value = 0
  transferPassword.value = ''
}

async function handleTransfer() {
  if (!transferTarget.value) return
  if (transferAmount.value <= 0) {
    appStore.showError(t('team.transfer.amountRequired'))
    return
  }
  if (!transferPassword.value) {
    appStore.showError(t('team.transfer.passwordRequired'))
    return
  }
  transferring.value = true
  try {
    await transferBalance(transferTarget.value.id, transferAmount.value, transferPassword.value)
    appStore.showSuccess(t('team.transfer.success'))
    closeTransfer()
    await loadMembers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.transfer.error')))
  } finally {
    transferring.value = false
  }
}

async function copyCode() {
  if (!team.value?.invite_code) return
  await copyToClipboard(team.value.invite_code, t('team.inviteCode.copied'))
}

function openDeposit() {
  fundMode.value = 'deposit'
  fundTarget.value = null
  fundAmount.value = 0
  fundPassword.value = ''
  fundModalOpen.value = true
}

function openAllocate(member: TeamMember) {
  fundMode.value = 'allocate'
  fundTarget.value = member
  fundAmount.value = 0
  fundPassword.value = ''
  fundModalOpen.value = true
}

function closeFund() {
  fundModalOpen.value = false
  fundTarget.value = null
  fundAmount.value = 0
  fundPassword.value = ''
}

async function handleFundSubmit() {
  if (!fundAmount.value || fundAmount.value <= 0) {
    appStore.showError(t('team.fund.amountRequired'))
    return
  }
  if (!fundPassword.value) {
    appStore.showError(t('team.fund.passwordRequired'))
    return
  }
  fundSubmitting.value = true
  try {
    if (fundMode.value === 'deposit') {
      await depositFund(fundAmount.value, fundPassword.value)
      appStore.showSuccess(t('team.fund.depositSuccess'))
    } else {
      if (!fundTarget.value) return
      await allocateFund(fundTarget.value.id, fundAmount.value, fundPassword.value)
      appStore.showSuccess(t('team.fund.allocateSuccess'))
    }
    closeFund()
    // 刷新团队资金余额 + 成员列表
    await loadTeam()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.fund.error')))
  } finally {
    fundSubmitting.value = false
  }
}
</script>
