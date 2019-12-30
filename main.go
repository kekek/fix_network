package main

func main()  {
	// 诊断 dns
		ips, err := net.LookupIP("google.com")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
			os.Exit(1)
		}
		for _, ip := range ips {
			fmt.Printf("google.com. IN A %s\n", ip.String())
		}

	// 诊断ip

}
