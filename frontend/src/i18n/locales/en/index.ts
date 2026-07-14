import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import admin from './admin'
import misc from './misc'
import playground from './playground'
import team from './team'
import tickets from './tickets'
import lottery from './lottery'

export default {
  ...landing,
  ...common,
  ...dashboard,
  admin,
  ...misc,
  ...playground,
  ...team,
  ...tickets,
  ...lottery,
}
