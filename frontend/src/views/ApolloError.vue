<template>
  <div class="error-bkg">
    <div class="error-message">
      <h3>We are sorry, an error has occurred!</h3>
      <div class="error-content">
        <h4>{{ message }}</h4>
        <p class="error-help">
          If you need further assistance, please contact Lou Foster (<a href="mailto:lf6f@virginia.edu">lf6f@virginia.edu</a>).
        </p>
        <p><router-link :to="{ name: 'home'}">Return to Collections List</router-link></p>
      </div>
    </div>
  </div>
</template>

<script>
  export default {
    props: {
      message: String
    }
  }
</script>

<style scoped>
  div.error-bkg {
    background-color: white;
    padding: 20px;
  }
  div.error-message {
    margin: 20px auto;
    border: 1px solid #a94442;
    width: 60%;
    background-color: white;
  }
  h3 {
    text-align: center;
    padding: 5px;
    margin: 0;
    border-bottom: 1px solid #a94442;
    background: #f2dede;
    color: #a94442;
  }
  div.error-content {
    padding:15px;
    text-align: center;
  }
</style>
