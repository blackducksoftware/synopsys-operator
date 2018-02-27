# Perceptor-protoform: a cloud native administration utility for your scanning platform.

Protoform is a cloud native installation utility for blackduck's distributed system framework for scanning platforms over time
for perceptor (core), perceptor-convex, perceivers.

Protoform can also be used by anyone in the community aiming to build a cloud
native installer for their applications that uses golang and native kubernetes
objects as a basis.

Protoform expects to run *inside a cluster* with privileges granted to it to 
*create replication controllers, services, and volumes*.

Depending on how you choose to run the perceptor project, hydra may also 
need to be running in a namespace where create options are allowed.

Note that, after hydra runs, you can remove any such service accounts, as they
are only needed for installation.

[1] http://tfwiki.net/wiki/Protoform
