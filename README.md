My repository for solving 1brc in Golang

I had a solution using bufio scanner (and no concurrency), but I didn't publish and deleted it.
It performed with time 3 minutes 3 seconds on my Macbook Air with M1 CPU

Now I redid this challenge by scanning bytes into a buffer myself. It performs worse at 3 minutes 31.8 seconds for the same process, but I will work on improving that.

Original 1brc challenge: https://github.com/gunnarmorling/1brc

This challenge is about optimizing read and calculation performance, not the end formatting. So I think it was fair for me to borrow result formatting and tests from this repo: https://github.com/shraddhaag/1brc
