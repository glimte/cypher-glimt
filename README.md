# Cypher-Glimt

Is a thin wrapper for Neo4j Cypher-Shell.    
It introduces the possibilities to use named profiles while working on multiple environments   
To get started with cypher-glimt, run it with command "cypher-glimt setup" to setup a named profile, address is specified in its full matter "bolt://172.16.0.1:7687"   
To reference a named profile, reference it with the --profile switch. Use --help or -h to view the help page.   
   
Future releases are planned to include password encryption and parsing of multiple files as input,   
for example by passing an entire folder using wildcards with the --file flag. 
