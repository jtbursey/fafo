# Ideas for added Functionality

- Allow for list of Ports
- Allow for more than one host
    - one cookie jar per host
- scan response body
    - html parser
    - xml parser (good for xmlrpc)
    - json parser
- ssl scanner
- php version detection
- pw sprayer (for more servers)
- scan common ports for the http server
    - automatically set port/http if not provided
        - test for both http and https if not specified
        - set port based on protocol
- runtime monitoring
    - debug timestamps
    - number of calls
- subdommain enum
- scan various TLS headers
- Signal repeated errors to quit
- scan robots.txt/humans.txt

<?xml version="1.0" encoding="utf-8"?><methodCall><methodName>wp.getUsersBlogs</methodName><params><param><value><string>admin</string></value></param><param><value><string>admin</string></value></param></params></methodCall>

<?xml version="1.0" encoding="utf-8"?><methodCall><methodName>wp.getProfile</methodName><params><param><value><int>1</int></value></param><param><value><string>admin</string></value></param><param><value><string>admin</string></value></param><param><value><array><data><value><string>user_id</string></value><value><string>display_name</string></value><value><string>email</string></value></data></array></value></param></params></methodCall>
