using CartService.Store;
using Grpc.Core;
using Grpc.Health.V1;
using Grpc.Net.Client;
using Hipstershop;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.AspNetCore.TestHost;
using Microsoft.Extensions.DependencyInjection;
using Xunit;

namespace CartService.Tests;

/// <summary>
/// Hosts the real cartservice (Kestrel pipeline via TestServer) with the
/// in-memory cart store, then drives it through generated gRPC clients —
/// a true smoke test of contract serialization, routing and health.
/// </summary>
public sealed class CartServiceAppFactory : WebApplicationFactory<Program>
{
    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        builder.ConfigureTestServices(services =>
            services.AddSingleton<ICartStore>(new LocalCartStore()));
    }
}

public sealed class CartServiceSmokeTests
{
    private static (Hipstershop.CartService.CartServiceClient cart, Health.HealthClient health) CreateClients()
    {
        var factory = new CartServiceAppFactory();
        var httpClient = factory.CreateClient();
        var channel = GrpcChannel.ForAddress(
            factory.Server.BaseAddress,
            new GrpcChannelOptions { HttpClient = httpClient });
        return (new Hipstershop.CartService.CartServiceClient(channel), new Health.HealthClient(channel));
    }

    [Fact]
    public async Task HealthCheck_NamedService_ReturnsServing()
    {
        var (_, health) = CreateClients();
        var response = await health.CheckAsync(new HealthCheckRequest { Service = "hipstershop.CartService" });
        Assert.Equal(HealthCheckResponse.Types.ServingStatus.Serving, response.Status);
    }

    [Fact]
    public async Task HealthCheck_Overall_ReturnsServing()
    {
        var (_, health) = CreateClients();
        var response = await health.CheckAsync(new HealthCheckRequest());
        Assert.Equal(HealthCheckResponse.Types.ServingStatus.Serving, response.Status);
    }

    [Fact]
    public async Task AddItem_Then_GetCart_ReturnsItem()
    {
        var (cart, _) = CreateClients();
        const string userId = "smoke-user-1";

        await cart.AddItemAsync(new AddItemRequest
        {
            UserId = userId,
            Item = new CartItem { ProductId = "OLJCESPC7Z", Quantity = 2 }
        });

        var got = await cart.GetCartAsync(new GetCartRequest { UserId = userId });
        Assert.Equal(userId, got.UserId);
        var item = Assert.Single(got.Items);
        Assert.Equal("OLJCESPC7Z", item.ProductId);
        Assert.Equal(2, item.Quantity);
    }

    [Fact]
    public async Task AddItem_SameProduct_MergesQuantity()
    {
        var (cart, _) = CreateClients();
        const string userId = "smoke-user-2";

        await cart.AddItemAsync(new AddItemRequest { UserId = userId, Item = new CartItem { ProductId = "P1", Quantity = 3 } });
        await cart.AddItemAsync(new AddItemRequest { UserId = userId, Item = new CartItem { ProductId = "P1", Quantity = 4 } });

        var got = await cart.GetCartAsync(new GetCartRequest { UserId = userId });
        var item = Assert.Single(got.Items);
        Assert.Equal(7, item.Quantity);
    }

    [Fact]
    public async Task GetCart_UnknownUser_ReturnsEmptyCart()
    {
        var (cart, _) = CreateClients();
        var got = await cart.GetCartAsync(new GetCartRequest { UserId = "nobody-here" });
        Assert.Equal("nobody-here", got.UserId);
        Assert.Empty(got.Items);
    }

    [Fact]
    public async Task EmptyCart_RemovesAllItems()
    {
        var (cart, _) = CreateClients();
        const string userId = "smoke-user-3";

        await cart.AddItemAsync(new AddItemRequest { UserId = userId, Item = new CartItem { ProductId = "A", Quantity = 1 } });
        await cart.AddItemAsync(new AddItemRequest { UserId = userId, Item = new CartItem { ProductId = "B", Quantity = 2 } });
        await cart.EmptyCartAsync(new EmptyCartRequest { UserId = userId });

        var got = await cart.GetCartAsync(new GetCartRequest { UserId = userId });
        Assert.Empty(got.Items);
    }

    [Theory]
    [InlineData("", "P1", 1)]   // empty user id
    [InlineData("u", "", 1)]    // empty product id
    [InlineData("u", "P1", 0)]  // zero quantity
    [InlineData("u", "P1", -1)] // negative quantity
    public async Task AddItem_InvalidArguments_ThrowsInvalidArgument(string userId, string productId, int quantity)
    {
        var (cart, _) = CreateClients();
        var ex = await Assert.ThrowsAsync<RpcException>(async () => await cart.AddItemAsync(new AddItemRequest
        {
            UserId = userId,
            Item = new CartItem { ProductId = productId, Quantity = quantity }
        }));
        Assert.Equal(StatusCode.InvalidArgument, ex.StatusCode);
    }
}
