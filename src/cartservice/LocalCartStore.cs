using System.Collections.Concurrent;
using Hipstershop;

namespace CartService.Store;

/// <summary>
/// Thread-safe in-memory cart store. Used when <c>CART_STORE=memory</c> or when
/// Redis is unreachable, so the service can start without any external dependency.
/// </summary>
public sealed class LocalCartStore : ICartStore
{
    private readonly ConcurrentDictionary<string, ConcurrentDictionary<string, int>> _carts =
        new(StringComparer.Ordinal);

    public Task AddItemAsync(string userId, string productId, int quantity, CancellationToken cancellationToken = default)
    {
        var items = _carts.GetOrAdd(userId, _ => new ConcurrentDictionary<string, int>(StringComparer.Ordinal));
        items.AddOrUpdate(productId, quantity, (_, existing) => checked(existing + quantity));
        return Task.CompletedTask;
    }

    public Task<Cart> GetCartAsync(string userId, CancellationToken cancellationToken = default)
    {
        var cart = new Cart { UserId = userId };
        if (_carts.TryGetValue(userId, out var items))
        {
            foreach (var pair in items)
            {
                cart.Items.Add(new CartItem { ProductId = pair.Key, Quantity = pair.Value });
            }
        }
        return Task.FromResult(cart);
    }

    public Task EmptyCartAsync(string userId, CancellationToken cancellationToken = default)
    {
        _carts.TryRemove(userId, out _);
        return Task.CompletedTask;
    }
}
