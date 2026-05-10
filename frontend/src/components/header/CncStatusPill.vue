<template>
  <router-link
    v-if="cnc.running && cnc.fileURL"
    :to="cnc.fileURL"
    class="cnc-pill"
    :title="t('cnc.statusTooltip', { path: cnc.filePath })"
  >
    <i class="material-icons cnc-pill__icon">precision_manufacturing</i>
    <span v-if="machineLabel" class="cnc-pill__machine">{{ machineLabel }}</span>
    <span class="cnc-pill__name">{{ basename(cnc.filePath) }}</span>
    <span class="cnc-pill__progress" v-if="cnc.lineTotal > 0">
      {{ cnc.lineCurrent }}&nbsp;/&nbsp;{{ cnc.lineTotal }}
    </span>
    <span class="cnc-pill__progress" v-else>
      {{ t("cnc.lineN", { n: cnc.lineCurrent }) }}
    </span>
  </router-link>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useCncStore } from "@/stores/cnc";

const { t } = useI18n();
const store = useCncStore();
const cnc = computed(() => store);

// Show the machine name as a prefix only when more than one machine
// is configured — single-machine setups already know which controller
// the pill refers to and the prefix is just visual noise.
const machineLabel = computed(() => {
  if (store.machines.length < 2) return "";
  return store.currentMachine?.name || store.currentMachineId || "";
});

const basename = (p: string) => {
  if (!p) return "";
  const parts = p.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? p;
};
</script>

<style>
.cnc-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.25rem 0.6rem;
  margin: 0 0.5rem;
  border-radius: 999px;
  background: var(--alt-background, rgba(33, 150, 243, 0.12));
  color: var(--primary-color, #2196f3);
  font-size: 0.85rem;
  font-weight: 500;
  text-decoration: none;
  white-space: nowrap;
  max-width: 50vw;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cnc-pill:hover {
  background: var(--alt-background, rgba(33, 150, 243, 0.2));
}

.cnc-pill__icon {
  font-size: 1.1rem;
}

.cnc-pill__name {
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 30vw;
}

.cnc-pill__machine {
  /* Shown only when >1 machine is configured. Slightly muted so the
     filename + progress remain the dominant read. */
  font-weight: 500;
  opacity: 0.85;
  padding-right: 0.4rem;
  border-right: 1px solid currentColor;
  border-right-color: rgba(33, 150, 243, 0.4);
  margin-right: 0.1rem;
}

.cnc-pill__progress {
  opacity: 0.7;
  font-variant-numeric: tabular-nums;
}
</style>
