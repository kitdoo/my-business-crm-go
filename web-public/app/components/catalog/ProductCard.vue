<script setup>
const props = defineProps({
  product: { type: Object, required: true },
})
const localePath = useLocalePath()
const imageUrl = computed(() => (props.product.imageIds?.[0] ? `/api/images/${props.product.imageIds[0]}` : '/images/product-placeholder.svg'))
</script>

<template>
  <NuxtLink :to="localePath(`/katalog/${product.sku}`)" class="block rounded-lg border border-black/10 overflow-hidden group">
    <div class="aspect-square overflow-hidden bg-gray-50">
      <NuxtImg :src="imageUrl" alt="" loading="lazy" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
    </div>
    <div class="p-4">
      <h3 class="font-medium mb-1 line-clamp-2"><LocalizedText :value="product.name" /></h3>
      <MoneyLabel v-if="product.price" :amount="product.price.amount" :currency="product.price.currency" class="font-semibold text-brand-700" />
      <div class="mt-2">
        <StatusBadge :status="product.status" />
      </div>
    </div>
  </NuxtLink>
</template>
