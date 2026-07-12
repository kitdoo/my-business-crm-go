<script setup>
const props = defineProps({
  product: { type: Object, required: true },
})
const { t } = useI18n()
const localePath = useLocalePath()
const { addItem, setQty, getQty, removeItem } = useCart()
const imageUrl = computed(() => (props.product.imageIds?.[0] ? `/api/images/${props.product.imageIds[0]}` : '/images/product-placeholder.svg'))
const qtyInCart = computed(() => getQty(props.product.sku))

function increment() {
  if (qtyInCart.value) setQty(props.product.sku, qtyInCart.value + 1)
  else addItem(props.product)
}
function decrement() {
  if (qtyInCart.value <= 1) removeItem(props.product.sku)
  else setQty(props.product.sku, qtyInCart.value - 1)
}
</script>

<template>
  <NuxtLink :to="localePath(`/katalog/${product.sku}`)" class="block rounded-lg border border-black/10 overflow-hidden group">
    <div class="aspect-square overflow-hidden bg-gray-50">
      <NuxtImg :src="imageUrl" alt="" loading="lazy" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
    </div>
    <div class="p-4">
      <h3 class="font-medium mb-1 line-clamp-2"><LocalizedText :value="product.name" /></h3>
      <MoneyLabel v-if="product.price" :amount="product.price.amount" :currency="product.price.currency" class="font-semibold text-brand-700" />
      <div class="mt-2 flex items-center justify-between gap-2">
        <StatusBadge :status="product.status" />

        <div v-if="qtyInCart" class="flex items-center gap-2" @click.stop.prevent>
          <button class="w-6 h-6 flex items-center justify-center rounded-full border border-brand-500 text-brand-700" @click="decrement">
            <UIcon name="i-lucide-minus" class="w-3 h-3" />
          </button>
          <span class="text-sm font-medium w-4 text-center">{{ qtyInCart }}</span>
          <button class="w-6 h-6 flex items-center justify-center rounded-full border border-brand-500 text-brand-700" @click="increment">
            <UIcon name="i-lucide-plus" class="w-3 h-3" />
          </button>
        </div>
        <button
          v-else
          class="text-xs uppercase tracking-wide text-brand-700 hover:text-brand-500 flex items-center gap-1"
          :aria-label="t('cart.addToCart')"
          @click.stop.prevent="addItem(product)"
        >
          <UIcon name="i-lucide-shopping-cart" class="w-4 h-4" />
        </button>
      </div>
    </div>
  </NuxtLink>
</template>
