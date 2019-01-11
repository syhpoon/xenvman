<template>
  <v-app>
    <v-toolbar app>
      <v-toolbar-title class="headline text-uppercase">
        <span class="font-weight-light">xenvman web ui</span>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <a href="https://github.com/syhpoon/xenvman/"
        target="_blank">
        <v-img
             :src="require('./assets/github.png')"
             contain
             height="32"
             width="32"
             class="my-3"/>
      </a>
    </v-toolbar>

    <v-content>
      <router-view/>
    </v-content>
  </v-app>
</template>

<script>

export default {
  name: 'App',
  data () {
    return {
    }
  }
}
</script>
