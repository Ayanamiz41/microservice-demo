using Grpc.Core;
using Grpc.Health.V1;

namespace CartService;

/// <summary>
/// Implementation of the standard <c>grpc.health.v1.Health</c> service.
/// </summary>
public sealed class HealthServiceImpl : Health.HealthBase
{
    public const string ServiceName = "hipstershop.CartService";

    public override Task<HealthCheckResponse> Check(HealthCheckRequest request, ServerCallContext context)
        => Task.FromResult(Respond(request.Service));

    private static HealthCheckResponse Respond(string? service) => new()
    {
        Status = string.IsNullOrEmpty(service) || service == ServiceName
            ? HealthCheckResponse.Types.ServingStatus.Serving
            : HealthCheckResponse.Types.ServingStatus.NotServing
    };
}
