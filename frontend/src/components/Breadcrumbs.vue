<template>
  <div class="breadcrumbs">
    <component
      :is="element"
      :to="base || ''"
      :aria-label="t('files.home')"
      :title="t('files.home')"
    >
      <i class="material-icons">home</i>
    </component>

    <span v-for="(link, index) in items" :key="index">
      <span class="chevron"
        ><i class="material-icons">keyboard_arrow_right</i></span
      >
      <component :is="element" :to="link.url">{{ link.name }}</component>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

const { t } = useI18n();

const route = useRoute();

const props = defineProps<{
  base: string;
  noLink?: boolean;
}>();

const buildPathBreadcrumbs = () => {
  const relativePath = route.path.replace(props.base, "");
  const parts = relativePath.split("/");

  if (parts[0] === "") {
    parts.shift();
  }

  if (parts[parts.length - 1] === "") {
    parts.pop();
  }

  const breadcrumbs: BreadCrumb[] = [];

  for (let i = 0; i < parts.length; i++) {
    const url =
      i === 0
        ? props.base + "/" + parts[i]
        : String(breadcrumbs[i - 1].url).replace(/\/$/, "") + "/" + parts[i];

    breadcrumbs.push({
      name: decodeURIComponent(parts[i]),
      url: url + "/",
    });
  }

  return breadcrumbs;
};

const items = computed(() => {
  const breadcrumbs = buildPathBreadcrumbs();
  const archiveQuery = Array.isArray(route.query.archive)
    ? route.query.archive[0]
    : route.query.archive;

  if (typeof archiveQuery === "string") {
    if (breadcrumbs.length > 0) {
      breadcrumbs[breadcrumbs.length - 1].url = {
        path: route.path,
        query: { archive: "/" },
      } as any;
    }

    const inner = archiveQuery.startsWith("/")
      ? archiveQuery.substring(1)
      : archiveQuery;
    const archiveParts = inner.split("/").filter(Boolean);
    let currentInner = "";

    for (const part of archiveParts) {
      currentInner += "/" + part;
      breadcrumbs.push({
        name: decodeURIComponent(part),
        url: {
          path: route.path,
          query: { archive: currentInner },
        } as any,
      });
    }
  }

  if (breadcrumbs.length > 3) {
    while (breadcrumbs.length !== 4) {
      breadcrumbs.shift();
    }

    breadcrumbs[0].name = "...";
  }

  return breadcrumbs;
});

const element = computed(() => {
  if (props.noLink) {
    return "span";
  }

  return "router-link";
});
</script>

<style></style>
