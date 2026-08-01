using CartService;
using CartService.Store;
using Grpc.Reflection;
using Microsoft.AspNetCore.Server.Kestrel.Core;
using StackExchange.Redis;

var builder = WebApplication.CreateBuilder(args);

// Default port 7070 (see src/README.md service matrix); override with PORT.
var port = int.TryParse(Environment.GetEnvironmentVariable("PORT"), out var p) ? p : 7070;
builder.WebHost.ConfigureKestrel(options =>
    options.ListenLocalhost(port, listen => listen.Protocols = HttpProtocols.Http2));

builder.Services.AddGrpc();
builder.Services.AddGrpcReflection();
builder.Services.AddSingleton<ICartStore>(CreateCartStore);
builder.Services.AddSingleton<CartServiceImpl>();
builder.Services.AddSingleton<HealthServiceImpl>();

var app = builder.Build();

app.MapGrpcService<CartServiceImpl>();
app.MapGrpcService<HealthServiceImpl>();
app.MapGrpcReflectionService();

app.Logger.LogInformation("cartservice listening on http://localhost:{Port}", port);
app.Run();

/// <summary>
/// Store selection: CART_STORE=memory forces the in-memory store;
/// otherwise Redis (REDIS_ADDR, default localhost:6379) is used, with an
/// automatic in-memory fallback so the service starts without dependencies.
/// </summary>
static ICartStore CreateCartStore(IServiceProvider sp)
{
    var logger = sp.GetRequiredService<ILoggerFactory>().CreateLogger("CartStore");
    var mode = (Environment.GetEnvironmentVariable("CART_STORE") ?? "redis").ToLowerInvariant();

    if (mode == "memory")
    {
        logger.LogInformation("Cart store: in-memory (CART_STORE=memory)");
        return new LocalCartStore();
    }

    var addr = Environment.GetEnvironmentVariable("REDIS_ADDR") ?? "localhost:6379";
    try
    {
        var mux = ConnectionMultiplexer.Connect(new ConfigurationOptions
        {
            EndPoints = { addr },
            AbortOnConnectFail = false,
            ConnectTimeout = 3000,
            SyncTimeout = 3000
        });
        mux.GetDatabase().Ping();
        logger.LogInformation("Cart store: Redis at {Addr}", addr);
        return new RedisCartStore(mux.GetDatabase(), mux);
    }
    catch (Exception ex)
    {
        logger.LogWarning("Redis unavailable at {Addr} ({Message}); falling back to in-memory cart store", addr, ex.Message);
        return new LocalCartStore();
    }
}

public partial class Program { }
