require 'tempfile'
require 'yaml'
require 'base64'

def dump_stream(array)
  array.map(&:to_yaml).join("\n")
end

def resources
  Dir.glob("kubernetes/*.yaml").flat_map { |f| YAML.load_stream(File.read(f)) }
end

desc "Run server locally and clean it up when done"
task :server do
  image = "kube_remediator"
  sh "docker build -t #{image} ."

  config = resources

  container = config.detect { |c| c["kind"] == "Deployment" }["spec"]["template"]["spec"]["containers"][0]

  # use the image we built
  container["image"] = image

  # do not pull images so our locally built image works
  container["imagePullPolicy"] = "Never"

  Tempfile.open(["kube_remediator", ".yaml"]) do |f|
    f.write dump_stream(config)
    f.close
    begin
      sh "kubectl apply -f #{f.path}"

      loop do
        puts "Waiting for server to start ..."
        sleep 1
        status = `kubectl get pods -l project=kube-remediator`
        break if status.include?("Running")

        if status.include?("CrashLoopBackOff")
          abort "Failed to start:\n" + `kubectl logs -l project=kube-remediator --previous`
        end
      end

      puts "Streaming logs ... Ctrl+c to shut down ..."
      sh "kubectl logs -l project=kube-remediator --follow"
    rescue Interrupt
      # user stopped ... do not print a backtrace
    ensure
      sh "kubectl delete -f #{f.path} --force --grace-period 0 --ignore-not-found"
    end
  end
end
