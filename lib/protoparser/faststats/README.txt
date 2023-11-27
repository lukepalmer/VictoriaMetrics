To generate the faststats codec:

Get the latest release of the SBE tooling:
https://github.com/real-logic/simple-binary-encoding/releases

Run the code generator as described here:
https://github.com/real-logic/simple-binary-encoding/wiki/Sbe-Tool-Guide

This looks like:
java -Dsbe.target.language=Golang -jar path/to/sbe-all-<version>.jar faststats.xml
