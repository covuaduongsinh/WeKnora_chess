<template>
  <div class="chess-manage">
    <h1 class="cm-title">Quản lý cờ vua</h1>
    <t-tabs v-model="tab" class="cm-tabs">
      <t-tab-panel value="courses" label="Khóa học">
        <div class="cm-pane"><ChessCourses /></div>
      </t-tab-panel>
      <t-tab-panel value="games" label="Kho ván đấu">
        <div class="cm-pane"><GameLibrary v-if="tab === 'games'" /></div>
      </t-tab-panel>
      <t-tab-panel value="puzzles" label="Ngân hàng bài tập">
        <div class="cm-pane"><PuzzleBank v-if="tab === 'puzzles'" /></div>
      </t-tab-panel>
    </t-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import ChessCourses from './ChessCourses.vue';
import GameLibrary from './GameLibrary.vue';
import PuzzleBank from './PuzzleBank.vue';

const tab = ref('courses');
</script>

<style lang="less" scoped>
.chess-manage { display: flex; flex-direction: column; height: 100%; box-sizing: border-box; padding: 12px 20px 0; overflow: hidden; }
.cm-title { font-size: 20px; margin: 0 0 8px; color: var(--td-text-color-primary); }
.cm-tabs { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.cm-tabs :deep(.t-tabs__content) { flex: 1; min-height: 0; }
.cm-tabs :deep(.t-tab-panel) { height: 100%; }
.cm-pane { height: 100%; min-height: 0; }
</style>
