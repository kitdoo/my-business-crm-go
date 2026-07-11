import brands from './brands.js'
import categories from './categories.js'
import clients from './clients.js'
import partners from './partners.js'
import users from './users.js'
import warehouses from './warehouses.js'

// Registry of all entities (TD §9.1). Add a new file + one import line
// here when a new entity ships — nothing else should need to know the
// full entity list.
const ENTITIES = {
  brands,
  categories,
  clients,
  partners,
  users,
  warehouses,
}

export function getEntityConfig(entityKey) {
  const config = ENTITIES[entityKey]
  if (!config) throw new Error(`Unknown entity: ${entityKey}`)
  return config
}

export function listEntityConfigs() {
  return Object.values(ENTITIES)
}
