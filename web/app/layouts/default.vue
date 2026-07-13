<script setup>
// App shell (TD §8): TopBar + SideMenu (flat menu, left). SideMenu is a
// permanent panel on >= lg and an overlay slideover below that (TD §6).
// Entity edit forms open in a right-side drawer from the list page itself
// (EntityListPage), not here.
const { menuOpen } = useLayoutState()
</script>

<template>
  <div class="h-screen overflow-hidden flex flex-col">
    <TopBar />
    <div class="flex flex-1 min-h-0">
      <aside class="hidden lg:block overflow-y-auto">
        <SideMenu />
      </aside>

      <main class="flex-1 min-w-0 p-4 overflow-y-auto">
        <slot />
      </main>
    </div>

    <USlideover v-model:open="menuOpen" side="left">
      <template #content>
        <SideMenu />
      </template>
    </USlideover>
  </div>
</template>
