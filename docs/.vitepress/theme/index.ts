import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
// @ts-ignore
import Layout from './Layout.vue'

export default {
    extends: DefaultTheme,
    Layout
} satisfies Theme
