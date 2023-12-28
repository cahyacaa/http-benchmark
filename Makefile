http-1:
	h2load -c10 -m250 -n10000 https://http1---korlantas-approver-tlmp6dxpfq-et.a.run.app/test

http-2:
	h2load -c10 -m250 -n10000 https://http2---korlantas-approver-tlmp6dxpfq-et.a.run.app/test