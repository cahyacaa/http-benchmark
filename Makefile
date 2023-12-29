http-1:
	h2load -c${conn} -m${multi} -n${req} https://http1---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-1mb

http-2:
	h2load -c${conn} -m${multi} -n${req} https://http2---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-1mb