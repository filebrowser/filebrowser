<template>
  <div>
    <div
      class="vue-simple-progress-text"
      :style="text_style"
      v-if="text.length > 0 && textPosition == 'top'"
    >
      {{ text }}
    </div>
    <div class="vue-simple-progress" :style="progress_style">
      <div
        class="vue-simple-progress-text"
        :style="text_style"
        v-if="text.length > 0 && textPosition == 'middle'"
      >
        {{ text }}
      </div>
      <div
        style="position: relative; left: -9999px"
        :style="text_style"
        v-if="text.length > 0 && textPosition == 'inside'"
      >
        {{ text }}
      </div>
      <div class="vue-simple-progress-bar" :style="bar_style">
        <div
          :style="text_style"
          v-if="text.length > 0 && textPosition == 'inside'"
        >
          {{ text }}
        </div>
      </div>
    </div>
    <div
      class="vue-simple-progress-text"
      :style="text_style"
      v-if="text.length > 0 && textPosition == 'bottom'"
    >
      {{ text }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const isNumber = (n: number | string): boolean => {
  return !isNaN(parseFloat(n as string)) && isFinite(n as number);
};

const props = withDefaults(
  defineProps<{
    val?: number;
    max?: number;
    size?: number | string;
    bgColor?: string;
    barColor?: string;
    barTransition?: string;
    barBorderRadius?: number;
    spacing?: number;
    text?: string;
    textAlign?: string;
    textPosition?: string;
    fontSize?: number;
    textFgColor?: string;
  }>(),
  {
    val: 0,
    max: 100,
    size: 3,
    bgColor: "#eee",
    barColor: "#2196f3",
    barTransition: "all 0.5s ease",
    barBorderRadius: 0,
    spacing: 4,
    text: "",
    textAlign: "center",
    textPosition: "bottom",
    fontSize: 13,
    textFgColor: "#222",
  }
);

const pct = computed(() => {
  const pct = (props.val / props.max) * 100;
  const pctFixed = pct.toFixed(2);
  return Math.min(parseFloat(pctFixed), props.max);
});

const size_px = computed(() => {
  switch (props.size) {
    case "tiny":
      return 2;
    case "small":
      return 4;
    case "medium":
      return 8;
    case "large":
      return 12;
    case "big":
      return 16;
    case "huge":
      return 32;
    case "massive":
      return 64;
  }

  return isNumber(props.size) ? (props.size as number) : 32;
});

const text_padding = computed(() => {
  switch (props.size) {
    case "tiny":
    case "small":
    case "medium":
    case "large":
    case "big":
    case "huge":
    case "massive":
      return Math.min(Math.max(Math.ceil(size_px.value / 8), 3), 12);
  }

  return isNumber(props.spacing) ? props.spacing : 4;
});

const text_font_size = computed(() => {
  switch (props.size) {
    case "tiny":
    case "small":
    case "medium":
    case "large":
    case "big":
    case "huge":
    case "massive":
      return Math.min(Math.max(Math.ceil(size_px.value * 1.4), 11), 32);
  }

  return isNumber(props.fontSize) ? props.fontSize : 13;
});

const progress_style = computed(() => {
  const style: Record<string, string> = {
    background: props.bgColor,
  };

  if (props.textPosition == "middle" || props.textPosition == "inside") {
    style["position"] = "relative";
    style["min-height"] = size_px.value + "px";
    style["z-index"] = "-2";
  }

  if (props.barBorderRadius > 0) {
    style["border-radius"] = props.barBorderRadius + "px";
  }

  return style;
});

const bar_style = computed(() => {
  const style: Record<string, string> = {
    background: props.barColor,
    width: pct.value + "%",
    height: size_px.value + "px",
    transition: props.barTransition,
  };

  if (props.barBorderRadius > 0) {
    style["border-radius"] = props.barBorderRadius + "px";
  }

  if (props.textPosition == "middle" || props.textPosition == "inside") {
    style["position"] = "absolute";
    style["top"] = "0";
    style["height"] = "100%";
    style["min-height"] = size_px.value + "px";
    style["z-index"] = "-1";
  }

  return style;
});

const text_style = computed(() => {
  const style: Record<string, string> = {
    color: props.textFgColor,
    "font-size": text_font_size.value + "px",
    "text-align": props.textAlign,
  };

  if (
    props.textPosition == "top" ||
    props.textPosition == "middle" ||
    props.textPosition == "inside"
  )
    style["padding-bottom"] = text_padding.value + "px";
  if (
    props.textPosition == "bottom" ||
    props.textPosition == "middle" ||
    props.textPosition == "inside"
  )
    style["padding-top"] = text_padding.value + "px";

  return style;
});
</script>
