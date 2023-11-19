import DefaultTheme from 'vitepress/theme'
import { Theme } from 'vitepress'
// @ts-ignore
import Layout from './Layout.vue'


export default {
    extends: DefaultTheme,
    Layout,
} as Theme
