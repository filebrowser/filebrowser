{
  # Name of our deployment
  network.description = "HelloWorld";
  # Enable rolling back to previous versions of our infrastructure
  network.enableRollback = true;

  # It consists of a single server named 'helloserver'
  helloserver =
    # Every server gets passed a few arguments, including a reference
    # to nixpkgs (pkgs)
    { config, pkgs, ... }:
    let
      # We import our custom packages from ./default passing pkgs as argument
      packages = import ./default.nix { pkgs = pkgs; };
      # This is the nodejs version specified in default.nix
      nodejs   = packages.nodejs;
      # And this is the application we'd like to deploy
      app      = packages.app;
    in
    {
      # We'll be running our application on port 8080, because a regular
      # user cannot bind to port 80
      # Then, using some iptables magic we'll forward traffic designated to port 80 to 8080
      networking.firewall.enable = true;
      # We will open up port 22 (SSH) as well otherwise we're locking ourselves out
      networking.firewall.allowedTCPPorts = [ 80 8080 22 ];
      networking.firewall.allowPing = true;

      # Port forwarding using iptables
      networking.firewall.extraCommands = ''
        iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
      '';

      # To run our node.js program we're going to use a systemd service
      # We can configure the service to automatically start on boot and to restart
      # the process in case it crashes
      systemd.services.helloserver = {
        description = "Hello world application";
        # Start the service after the network is available
        after = [ "network.target" ];
        # We're going to run it on port 8080 in production
        environment = { PORT = "8080"; };
        serviceConfig = {
          # The actual command to run
          ExecStart = "${nodejs}/bin/node ${app}/server.js";
          # For security reasons we'll run this process as a special 'nodejs' user
          User = "nodejs";
          Restart = "always";
        };
      };

      # And lastly we ensure the user we run our application as is created
      users.extraUsers = {
        nodejs = { };
      };
    };
}