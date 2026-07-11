// Strips the response down to the fields the browser is allowed to see
// (TD §4.1 step 2) — never forward token, and only the public User fields.
export function publicUser(user) {
  if (!user) return null
  return {
    id: user.id,
    name: user.name,
    lastName: user.lastName,
    phone: user.phone,
    email: user.email,
    role: user.role,
    status: user.status,
  }
}
