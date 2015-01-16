# Blasted Media Server

The BMS is a simple media server designed to serve assets of various types.
Examples of assets are:
* Images
* Video files
* Journal entries
* Blog posts
* Movies
* TV Shows
* Books

Assets have meta data, and that metadata is specified by the asset type.  For
instance, a Movie asset may have metadata indicating the studio that produced
the movie, the list of lead actors as well as a plot synopsis.

All assets belong to one ore more collections, and a collection of assets can
be contained within another collection.  A given instance of the BMS has a root
collection.  Top level collections can live in the root collection, but not
assets directly.

Collection Layout Example:
```
├── Movies
│   ├── Avatar
│   ├── Spies Like Us
│   └── Star Trek
│       ├── Star Trek: The Motion Picture (1979)
│       ├── Star Trek II The Wrath of Khan (1982)
│       ├── Star Trek III The Search for Spock (1984)
│       ├── Star Trek IV The Voyage Home (1986)
│       ├── Star Trek V: The Final Frontier (1989)
│       ├── Star Trek VI The Undiscovered Country (1991)
│       ├── Star Trek: Generations (1994)
│       ├── Star Trek First Contact (1996)
│       ├── Star Trek: Insurrection (1998)
│       ├── Star Trek Nemesis (2002)
│       ├── Star Trek (2009)
│       └─── Star Trek Into Darkness (2013)
└── TV Shows
    ├── House
    │   ├── Season 1
    │   ├── Season 2
    │   ├── Season 3
    │   ├── Season 4
    │   ├── Season 5
    │   ├── Season 6
    │   ├── Season 7
    │   └── Season 8
    ├── MASH
    └── Police Squad
        └── Season 1

```

The BMS also includes an authorization system of users and groups.  A user
belongs to zero or more groups.  A group is used to change the assets available
to a user.  Groups change the available assets using filters.

A filter can specify various matches and conditionals for the 

