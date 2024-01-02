http-1:
	h2load -c${conn} -m${multi} -n${req} --h1 https://http1---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-1mb

http-2:
	h2load -c${conn} -m${multi} -n${req} https://http2---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-1mb

http-1:
	h2load -m POST -c${conn} -m${multi} -n${req} https://http1---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-1mb -d testpayload.json --h1 --header 'Content-Type: application/json'

http-2:
	h2load -m POST -c${conn} -m${multi} -n${req} https://http2---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-1mb -d testpayload.json --header 'Content-Type: application/json'