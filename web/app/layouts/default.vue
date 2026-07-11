<script setup>
// App shell (TD §8): TopBar + RightDrawer (flat menu). RightDrawer is a
// permanent panel on >= lg and an overlay slideover below that (TD §6).
// Entity edit forms open in a left-side drawer from the list page itself
// (EntityListPage), not here.
const { rightDrawerOpen } = useLayoutState()
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <TopBar />
    <div class="flex flex-1 min-h-0">
      <main class="flex-1 min-w-0 p-4 overflow-y-auto">
        <slot />
      </main>

      <aside class="hidden lg:block">
        <RightDrawer />
      </aside>
    </div>

    <USlideover v-model:open="rightDrawerOpen" side="right">
      <template #content>
        <RightDrawer />
      </template>
    </USlideover>
  </div>
</template>
