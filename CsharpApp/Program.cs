using System.IO;
using System;
using System.Net;
using System.Net.Http;
using System.Net.Security;
using System.Reflection;
using System.Security.Authentication;
using System.Security.Cryptography.X509Certificates;
using System.Threading.Tasks;
namespace CsharpApp
{
    class Program
    {
         static  async Task Main(string[] args)
        {
            var handler = new HttpClientHandler()
            {
                SslProtocols = SslProtocols.Tls12,
                ClientCertificateOptions = ClientCertificateOption.Manual,
                ServerCertificateCustomValidationCallback = (message, cer, chain, errors) =>
                {
                    return chain.Build(cer);
                }
            };
            var path =AppDomain.CurrentDomain.BaseDirectory + "cert\\client.pfx";
            var crt = new X509Certificate2(path, "123456789");
            handler.ClientCertificates.Add(crt);

           var client = new HttpClient(handler){DefaultRequestVersion = new Version(2, 0)}; ;

            var url = "https://localhost:8080/";
            var response = await client.GetAsync(url);
            var back = await response.Content.ReadAsStringAsync();
            Console.WriteLine(back);
            
        }
    }
}
