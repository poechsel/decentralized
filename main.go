package main
import "flag"

func main() {
    var _ = flag.String("UIPort", "8080", "Port for the UI client")
    var _ = flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
    var _ = flag.String("name", "", "name of the gossiper")
    var _ = flag.String("peers", "", "comma separated list of peers of the form ip:port")
    var _ = flag.Bool("simple", false, "run gossiper in simple broadcast mode")
    flag.Parse()
}


