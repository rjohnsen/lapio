## Configuration file

In order to connect to ElasticSearch you need create the file **elastic.toml** in current directory where you run this tool. 
Then use the following as template for your connection: 

```toml
host = "http://<ELASTIC IP HERE>:9200"
username = "elastic"
password = ""
workers = 10
flushbytes = 5
flushinterval = 30
```

## Parser Directive

A parser directive is as simple as an XML file describing how a log file is to be parsed. It contains regexes and other mapping information. 

### Pro hints and tips

* Use named capture groups in regex (Golang flavored)