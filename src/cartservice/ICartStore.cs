using Hipstershop;

namespace CartService.Store;

/// <summary>
/// Storage abstraction for shopping carts.
/// Implementations: <see cref="RedisCartStore"/> (default) and
/// <see cref="LocalCartStore"/> (in-memory, dependency-free fallback).
/// </summary>
public interface ICartStore
{
    /// <summary>Adds <paramref name="quantity"/> of <paramref name="productId"/> to the user's cart, merging with any existing quantity.</summary>
    Task AddItemAsync(string userId, string productId, int quantity, CancellationToken cancellationToken = default);

    /// <summary>Returns the user's cart (empty cart when the user has no cart yet).</summary>
    Task<Cart> GetCartAsync(string userId, CancellationToken cancellationToken = default);

    /// <summary>Removes the user's entire cart.</summary>
    Task EmptyCartAsync(string userId, CancellationToken cancellationToken = default);
}
