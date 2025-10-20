<template>
  <div>
    <h2>Products</h2>
    <div v-if="loading">Loading...</div>
    <div v-else>
      <div v-for="product in products" :key="product.id" class="product">
        <span>{{ product.name }}</span>
        <input type="number" v-model.number="product.quantity" min="1">
        <button @click="addToCart(product)">Add to Cart</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      products: [],
      loading: true,
    };
  },
  methods: {
    addToCart(product) {
      this.$emit('add-to-cart', product);
    },
  },
  mounted() {
    fetch('/api/products')
      .then(response => response.json())
      .then(data => {
        this.products = data.map(product => ({ ...product, quantity: 1 }));
        this.loading = false;
      })
      .catch(error => {
        console.error('Error fetching products:', error);
        this.loading = false;
      });
  },
};
</script>

<style scoped>
.product {
  margin-bottom: 10px;
}
</style>
