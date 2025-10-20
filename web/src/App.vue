<template>
  <div id="app">
    <h1>My Store</h1>
    <ProductList @add-to-cart="addToCart" />
    <Cart :cart="cart" @order-submitted="clearCart" />
  </div>
</template>

<script>
import ProductList from './components/ProductList.vue';
import Cart from './components/Cart.vue';

export default {
  name: 'App',
  components: {
    ProductList,
    Cart,
  },
  data() {
    return {
      cart: [],
    };
  },
  methods: {
    addToCart(product) {
      const existingProduct = this.cart.find(item => item.id === product.id);
      if (existingProduct) {
        existingProduct.quantity += product.quantity;
      } else {
        this.cart.push({ ...product });
      }
    },
    clearCart() {
      this.cart = [];
    },
  },
};
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
