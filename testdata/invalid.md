An invalid diagram falls back to its source, escaped:

```d2
<script>alert("xss")</script> & <img src=x onerror=alert(1)>
```
