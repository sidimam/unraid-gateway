import json
from http.server import BaseHTTPRequestHandler, HTTPServer
class H(BaseHTTPRequestHandler):
    def do_POST(self):
        n=int(self.headers.get('content-length') or 0); self.rfile.read(n)
        key=self.headers.get('x-api-key')
        if self.path!='/graphql' or key!='good-key':
            self.send_response(401); self.end_headers(); self.wfile.write(b'{"errors":[{"message":"unauthorized"}]}'); return
        body=json.dumps({"data":{"me":{"id":"1","name":"root","roles":["admin"]}}}).encode()
        self.send_response(200); self.send_header('Content-Type','application/json'); self.end_headers(); self.wfile.write(body)
    def log_message(self,*a): pass
HTTPServer(('0.0.0.0',18080),H).serve_forever()
