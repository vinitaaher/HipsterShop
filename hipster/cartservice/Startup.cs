using System;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using OpenTelemetry;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using cartservice.cartstore;
using cartservice.services;
using System.Collections;
using System.Collections.Generic;


namespace cartservice
{
    public class Startup
    {
        public Startup(IConfiguration configuration)
        {
            Configuration = configuration;
        }

        public IConfiguration Configuration { get; }

        // This method gets called by the runtime. Use this method to add services to the container.
        // For more information on how to configure your application, visit https://go.microsoft.com/fwlink/?LinkID=398940
        public void ConfigureServices(IServiceCollection services)
        {
            string redisAddress = Configuration["REDIS_ADDR"];
            ICartStore cartStore = null;
            if (!string.IsNullOrEmpty(redisAddress))
            {
                cartStore = new RedisCartStore(redisAddress);
            }
            else
            {
                Console.WriteLine("Redis cache host(hostname+port) was not specified. Starting a cart service using local store");
                Console.WriteLine("If you wanted to use Redis Cache as a backup store, you should provide its address via command line or REDIS_ADDRESS environment variable.");
                cartStore = new LocalCartStore();
            }

            // Initialize the redis store
            cartStore.InitializeAsync().GetAwaiter().GetResult();
            Console.WriteLine("Initialization completed");

            services.AddControllers();
            services.AddGrpc();
            services.AddSingleton<ICartStore>(cartStore);
            services.AddOpenTelemetryTracing(builder => ConfigureOpenTelemetry(builder, cartStore));
       

        }
        private static void ConfigureOpenTelemetry(TracerProviderBuilder builder, ICartStore cartStore)
        {
            builder.AddAspNetCoreInstrumentation();

            if (cartStore is RedisCartStore redisCartStore)
            {
                builder.AddRedisInstrumentation(redisCartStore.ConnectionMultiplexer);
            }
       
            var exportType = Environment.GetEnvironmentVariable("EXPORT_TYPE") ?? "JAEGER";
            var serviceName = Environment.GetEnvironmentVariable("SERVICE_NAME") ?? "CARTSERVICE"+ (exportType == "JAEGER" ? string.Empty : $"-{exportType}");
            var myList = new List< KeyValuePair<string, object>>();
            myList.Add(new KeyValuePair<string, object>("host.name", Environment.GetEnvironmentVariable("HOST_NAME")));
            myList.Add(new KeyValuePair<string, object>("resource.type", Environment.GetEnvironmentVariable("RESOURCE_TYPE")));
            myList.Add(new KeyValuePair<string, object>("ip",Environment.GetEnvironmentVariable("MY_POD_IP")));
            builder.SetResourceBuilder(ResourceBuilder.CreateDefault().AddService(serviceName,Environment.GetEnvironmentVariable("SERVICE_NAMESPACE"), null, false, $"{serviceName}-{exportType}-{Guid.NewGuid().ToString()}").AddAttributes(myList));
              switch (exportType)
            {
                case "OTLP":
                    var otlpEndpoint =Environment.GetEnvironmentVariable("OTLP_ENDPOINT") ?? "localhost:55680";
                    builder
                        .AddOtlpExporter(options => options.Endpoint = otlpEndpoint);
                    break;
                case "JAEGER":
                default:

                    var jaegerEndpoint = Environment.GetEnvironmentVariable("JAEGER_ENDPOINT") ?? "http://localhost:14268/api/traces";
                    builder.AddAspNetCoreInstrumentation();
                    var agenthost = Environment.GetEnvironmentVariable("AgentHost") ?? "localhost" ;
                    JaegerExporterHelperExtensions.AddJaegerExporter(builder,options => options.AgentHost =agenthost );
                    Console.WriteLine("Jaeger Tracing completed");
                    break;
            }
        }
        // This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
        public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
            }

            app.UseRouting();

            app.UseEndpoints(endpoints =>
            {
                endpoints.MapGrpcService<CartService>();
                endpoints.MapGrpcService<cartservice.services.HealthCheckService>();

                endpoints.MapGet("/", async context =>
                {
                    await context.Response.WriteAsync("Communication with gRPC endpoints must be made through a gRPC client. To learn how to create a client, visit: https://go.microsoft.com/fwlink/?linkid=2086909");
                });
            });
        }
    }
    }
    
