<script setup lang="ts">
import DefaultTheme from 'vitepress/theme'
import './style.css'
</script>

<template>
  <DefaultTheme.Layout>
    <template #home-hero-image>
      <div class="hero-terminal" aria-hidden="true">
        <div class="terminal-top">
          <span></span><span></span><span></span>
          <b>onebot / bendan</b>
        </div>
        <div class="terminal-body">
          <p><i>$</i> bendan status</p>
          <p><strong>connected</strong> ws://127.0.0.1:3001</p>
          <p><i>+</i> 33 zh actions</p>
          <p><i>+</i> 12 latin aliases</p>
          <p class="cursor"><i>$</i> _</p>
        </div>
      </div>
    </template>
  </DefaultTheme.Layout>
</template>
