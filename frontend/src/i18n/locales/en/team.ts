export default {
  team: {
    title: 'My Team',
    description: 'Create or join a team to share balance and view usage with members',
    loadError: 'Failed to load team information',
    role: {
      owner: 'Owner',
      member: 'Member'
    },
    create: {
      title: 'Create Team',
      description: 'Become an owner and invite members',
      namePlaceholder: 'Team name',
      button: 'Create Team',
      creating: 'Creating...',
      nameRequired: 'Please enter a team name',
      success: 'Team created successfully',
      error: 'Failed to create team'
    },
    join: {
      title: 'Join Team',
      description: 'Enter the invite code from the team owner',
      codePlaceholder: 'Invite code',
      button: 'Join Team',
      joining: 'Joining...',
      codeRequired: 'Please enter an invite code',
      success: 'Joined team successfully',
      error: 'Failed to join team'
    },
    inviteCode: {
      refresh: 'Refresh Code',
      refreshSuccess: 'Invite code refreshed',
      refreshError: 'Failed to refresh invite code',
      copied: 'Invite code copied'
    },
    leave: {
      button: 'Leave Team',
      leaving: 'Leaving...',
      confirm: 'Are you sure you want to leave the team?',
      success: 'Left team successfully',
      error: 'Failed to leave team'
    },
    members: {
      title: 'Team Members',
      loadError: 'Failed to load members',
      empty: 'No team members yet',
      searchPlaceholder: 'Search member email or username',
      balanceHidden: 'Hidden',
      usageHidden: 'Hidden',
      remove: 'Remove',
      removeConfirm: 'Are you sure you want to remove member {email}?',
      removeSuccess: 'Member removed',
      removeError: 'Failed to remove member',
      count: '{count} members',
      columns: {
        user: 'User',
        balance: 'Balance',
        usage: 'Total Usage',
        actions: 'Actions'
      },
      usage: {
        startDate: 'Start Date',
        endDate: 'End Date',
        query: 'Query Usage',
        empty: 'No usage records in this period',
        loadError: 'Failed to load usage records',
        noPermission: 'No permission to view',
        time: 'Time',
        model: 'Model',
        type: 'Type',
        tokens: 'Tokens',
        cost: 'Cost',
        duration: 'Duration'
      }
    },
    fund: {
      balance: 'Team Fund',
      depositButton: 'Deposit',
      depositTitle: 'Deposit to Team Fund',
      depositHint: 'Move funds from your personal balance into the team fund',
      allocateButton: 'Allocate',
      allocateTitle: 'Allocate Team Fund',
      allocateTo: 'Allocate team fund to {email}',
      amount: 'Amount',
      amountPlaceholder: 'Enter amount',
      password: 'Confirm Login Password',
      passwordPlaceholder: 'Enter your login password',
      confirm: 'Confirm',
      submitting: 'Processing...',
      depositSuccess: 'Deposited to team fund',
      allocateSuccess: 'Fund allocated successfully',
      error: 'Operation failed',
      amountRequired: 'Please enter a valid amount',
      passwordRequired: 'Please enter your login password'
    },
    transfer: {
      button: 'Transfer Balance',
      title: 'Transfer Balance',
      to: 'Transfer balance to {email}',
      amount: 'Amount',
      amountPlaceholder: 'Enter amount',
      password: 'Password Confirmation',
      passwordPlaceholder: 'Enter your login password',
      confirm: 'Confirm Transfer',
      transferring: 'Transferring...',
      amountRequired: 'Please enter a valid amount',
      passwordRequired: 'Please enter your password',
      success: 'Balance transferred successfully',
      error: 'Failed to transfer balance'
    }
  },
}
