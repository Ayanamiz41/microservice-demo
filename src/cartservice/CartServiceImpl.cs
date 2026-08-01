using CartService.Store;
using Grpc.Core;
using Hipstershop;

namespace CartService;

/// <summary>
/// Implementation of the <c>hipstershop.CartService</c> contract
/// (AddItem / GetCart / EmptyCart), aligned with the upstream
/// Online Boutique cartservice.
/// </summary>
public sealed class CartServiceImpl : Hipstershop.CartService.CartServiceBase
{
    private readonly ICartStore _store;

    public CartServiceImpl(ICartStore store) => _store = store;

    public override async Task<Empty> AddItem(AddItemRequest request, ServerCallContext context)
    {
        if (string.IsNullOrEmpty(request.UserId)
            || request.Item is null
            || string.IsNullOrEmpty(request.Item.ProductId)
            || request.Item.Quantity <= 0)
        {
            throw new RpcException(new Status(StatusCode.InvalidArgument, "UserId, ProductId and Quantity are required."));
        }

        try
        {
            await _store.AddItemAsync(request.UserId, request.Item.ProductId, request.Item.Quantity, context.CancellationToken);
        }
        catch (Exception ex)
        {
            throw new RpcException(new Status(StatusCode.Internal, ex.Message));
        }

        return new Empty();
    }

    public override async Task<Cart> GetCart(GetCartRequest request, ServerCallContext context)
    {
        if (string.IsNullOrEmpty(request.UserId))
        {
            throw new RpcException(new Status(StatusCode.InvalidArgument, "UserId is required."));
        }

        try
        {
            return await _store.GetCartAsync(request.UserId, context.CancellationToken);
        }
        catch (Exception ex)
        {
            throw new RpcException(new Status(StatusCode.Internal, ex.Message));
        }
    }

    public override async Task<Empty> EmptyCart(EmptyCartRequest request, ServerCallContext context)
    {
        if (string.IsNullOrEmpty(request.UserId))
        {
            throw new RpcException(new Status(StatusCode.InvalidArgument, "UserId is required."));
        }

        try
        {
            await _store.EmptyCartAsync(request.UserId, context.CancellationToken);
        }
        catch (Exception ex)
        {
            throw new RpcException(new Status(StatusCode.Internal, ex.Message));
        }

        return new Empty();
    }
}
