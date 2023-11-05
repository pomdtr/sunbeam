<script setup>
import DefaultTheme from 'vitepress/theme'

const { Layout } = DefaultTheme
</script>

<template>
  <Layout>
    <template #home-hero-image>
        <video autoplay muted loop style="border-radius: 10px;">
            <source src="/assets/demo.webm" type="video/webm">
        </video>
    </template>
  </Layout>
</template>
