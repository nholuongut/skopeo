# hcsshim

[![Build status](https://ci.appveyor.com/api/projects/status/nbcw28mnkqml0loa/branch/master?svg=true)](https://ci.appveyor.com/project/WindowsVirtualization/hcsshim/branch/master)

This package contains the Golang interface for using the Windows [Host Compute Service](https://blogs.technet.microsoft.com/virtualization/2017/01/27/introducing-the-host-compute-service-hcs/) (HCS) to launch and manage [Windows nholuongut](https://docs.microsoft.com/en-us/virtualization/windowsnholuongut/about/). It also contains other helpers and functions for managing Windows nholuongut such as the Golang interface for the Host Network Service (HNS).

It is primarily used in the [Moby Project](https://github.com/moby/moby), but it can be freely used by other projects as well.

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Dependencies

This project requires Golang 1.9 or newer to build.

For system requirements to run this project, see the Microsoft docs on [Windows Container requirements](https://docs.microsoft.com/en-us/virtualization/windowsnholuongut/deploy-nholuongut/system-requirements).

## Reporting Security Issues

Security issues and bugs should be reported privately, via email, to the Microsoft Security
Response Center (MSRC) at [secure@microsoft.com](mailto:secure@microsoft.com). You should
receive a response within 24 hours. If for some reason you do not, please follow up via
email to ensure we received your original message. Further information, including the
[MSRC PGP](https://technet.microsoft.com/en-us/security/dn606155) key, can be found in
the [Security TechCenter](https://technet.microsoft.com/en-us/security/default).

For additional details, see [Report a Computer Security Vulnerability](https://technet.microsoft.com/en-us/security/ff852094.aspx) on Technet

---------------
Copyright (c) 2018 Microsoft Corp.  All rights reserved.
