# LED-CLI

Link Extractor and Downloader

Run from the command line and queries the user to select one of the files found
in the current directory that match html, html, or eml extension. Parses that
file for any anchor tags in the html that match href and if it is a link to an
http\* site, it downloads the target.

---

A friend of mine had a problem at work where one of the accountants was getting
emailed a couple times a month with up to 200 links to PDF invoices. They had
to click each one to download them so I figured that it might be a good
exercise to write a program to solve that problem.

This was an attempt for me to also learn the Go programming language as I have
only written very small programs with it for experimentation. I have greatly
enjoyed the work so far and for the most part, the functionality is complete.
I still would like to make it more abstract in the way it behaves so that it can
possible be used for more than just PDF's in an email. I also still would like
to utilize Go Interfaces but have not had much of a chance in this program but
may try to rework things in the future to utilize one.
