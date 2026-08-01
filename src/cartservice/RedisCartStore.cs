using Hipstershop;
using StackExchange.Redis;

namespace CartService.Store;

/// <summary>
/// Redis-backed cart store (aligned with the upstream Online Boutique cartservice).
/// One Redis hash per user, key <c>cart:&lt;userId&gt;</c>, field = product id,
/// value = quantity.
/// </summary>
public sealed class RedisCartStore : ICartStore, IDisposable
{
    private const string KeyPrefix = "cart:";

    private readonly IDatabase _db;
    private readonly ConnectionMultiplexer? _mux;

    public RedisCartStore(IDatabase db, ConnectionMultiplexer? mux = null)
    {
        _db = db;
        _mux = mux;
    }

    public Task AddItemAsync(string userId, string productId, int quantity, CancellationToken cancellationToken = default)
        => _db.HashIncrementAsync(KeyPrefix + userId, productId, quantity);

    public async Task<Cart> GetCartAsync(string userId, CancellationToken cancellationToken = default)
    {
        var entries = await _db.HashGetAllAsync(KeyPrefix + userId);
        var cart = new Cart { UserId = userId };
        foreach (var entry in entries)
        {
            if (entry.Value.TryParse(out int quantity) && quantity > 0)
            {
                cart.Items.Add(new CartItem { ProductId = entry.Name, Quantity = quantity });
            }
        }
        return cart;
    }

    public Task EmptyCartAsync(string userId, CancellationToken cancellationToken = default)
        => _db.KeyDeleteAsync(KeyPrefix + userId);

    public void Dispose() => _mux?.Dispose();
}
