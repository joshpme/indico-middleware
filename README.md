
# Indico Middleware

I have setup a number of scripts to fetch various bits of data from indico and consolidate them into a mongodb instance.

A user can then query the mongodb instance to get the data they need.

For example:

https://faas-syd1-c274eac6.doserverless.co/api/v1/web/fn-19977d5d-a466-4a2d-bfd5-e29ba32197eb/indico/find?conference=41&code=TUPA071

Input is `conference: 41, code: TUPA071`

Output is the paper details for that conference, including the authors the order they should appear on the paper and the affiliations.