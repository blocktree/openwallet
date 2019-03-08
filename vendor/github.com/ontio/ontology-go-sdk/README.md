
<h1 align="center">Go SDK For Ontology  </h1>
<h4 align="center">Version 0.6.0 </h4>

## Overview
This is a comprehensive Go library for the Ontology blockchain. Currently, it supports local wallet management, digital asset management,  deployment and invoke for Smart Contract , and communication with Ontology Blockchain. The future will also support more rich functions and applications .

## How to use?

First of all, Create OntologySDK instance by NewOntologySdk method.

`sdk := NewOntologySdk()`

Then, set rpc server address.

`sdk.Rpc.SetAddress("http://localhost:20336")`

Then, call rpc server through sdk instance.

`sdk.Rpc.GetVersion()`


# Contributing

Can I contribute patches to Ontology project?

Yes! Please open a pull request with signed-off commits. We appreciate your help!

You can also send your patches as emails to the developer mailing list.
Please join the Ontology mailing list or forum and talk to us about it.

Either way, if you don't sign off your patches, we will not accept them.
This means adding a line that says "Signed-off-by: Name <email>" at the
end of each commit, indicating that you wrote the code and have the right
to pass it on as an open source patch.

Also, please write good git commit messages.  A good commit message
looks like this:

  Header line: explain the commit in one line (use the imperative)

  Body of commit message is a few lines of text, explaining things
  in more detail, possibly giving some background about the issue
  being fixed, etc etc.

  The body of the commit message can be several paragraphs, and
  please do proper word-wrap and keep columns shorter than about
  74 characters or so. That way "git log" will show things
  nicely even when it's indented.

  Make sure you explain your solution and why you're doing what you're
  doing, as opposed to describing what you're doing. Reviewers and your
  future self can read the patch, but might not understand why a
  particular solution was implemented.

  Reported-by: whoever-reported-it
  Signed-off-by: Your Name <youremail@yourhost.com>

## Website

* https://ont.io/

## License

The Ontology library (i.e. all code outside of the cmd directory) is licensed under the GNU Lesser General Public License v3.0, also included in our repository in the License file.

