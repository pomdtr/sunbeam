<script setup>
import DefaultTheme from 'vitepress/theme'

const { Layout } = DefaultTheme
</script>

<template>
  <Layout>
    <template #home-hero-image>
        <div class="demo-wrapper">
        <img src="https://vhs.charm.sh/vhs-1331psn9qEJqr4QhiCMSMd.gif" class="demo" />
        </div>
    </template>
  </Layout>
</template>

<style>
    .demo-wrapper {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 100%; /* Adjust the height as needed to match the parent div's height */
    }

    .demo {
        margin: auto;
        border: double 4px transparent;
        border-radius: 10px;
        background-image: linear-gradient(white, white), radial-gradient(circle at top left, #f00, #3020ff);
        background-origin: border-box;
        background-clip: padding-box, border-box;
    }
</style>
