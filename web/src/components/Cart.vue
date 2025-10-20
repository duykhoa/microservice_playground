<template>
  <div>
    <h2>Cart</h2>
    <div v-if="cart.length === 0">Cart is empty</div>
    <div v-else>
      <div v-for="item in cart" :key="item.id" class="cart-item">
        <span>{{ item.name }} ({{ item.quantity }})</span>
      </div>
      <button @click="submitOrder">Submit Order</button>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    cart: {
      type: Array,
      required: true,
    },
  },
  methods: {
    submitOrder() {
      fetch('/api/orders', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ items: this.cart }),
      })
        .then(response => {
          if (response.ok) {
            alert('Order submitted successfully!');
            this.$emit('order-submitted');
          } else {
            alert('Failed to submit order.');
          }
        })
        .catch(error => {
          console.error('Error submitting order:', error);
          alert('Failed to submit order.');
        });
    },
  },
};
</script>

<style scoped>
.cart-item {
  margin-bottom: 10px;
}
</style>
