// Stub submission (TZ §7.1 variant A/B — decision deferred for this
// iteration). Swap the body of these two functions for a real email send
// or a CRM lead/Partner creation once that decision is made; the route
// handlers and the client-side forms do not need to change either way.

export async function submitDealerForm(data) {
  console.info('[forms] dealer submission (stub, not sent anywhere yet):', data)
  return { ok: true }
}

export async function submitContactForm(data) {
  console.info('[forms] contact submission (stub, not sent anywhere yet):', data)
  return { ok: true }
}
