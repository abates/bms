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

Assets belong to folders and folders can have one ore more sub-folders.  The
system has a root folder that top-level folders belong to.  However, assets can
not be assigned to the root folder.

Folder Layout Example:
```
├── Movies
│   ├── Avatar
│   ├── Spies Like Us
│   ├── Star Trek: The Motion Picture (1979)
│   ├── Star Trek II The Wrath of Khan (1982)
│   ├── Star Trek III The Search for Spock (1984)
│   ├── Star Trek IV The Voyage Home (1986)
│   ├── Star Trek V: The Final Frontier (1989)
│   ├── Star Trek VI The Undiscovered Country (1991)
│   ├── Star Trek: Generations (1994)
│   ├── Star Trek First Contact (1996)
│   ├── Star Trek: Insurrection (1998)
│   ├── Star Trek Nemesis (2002)
│   ├── Star Trek (2009)
│   └── Star Trek Into Darkness (2013)
└── TV Shows
    ├── House
    ├── MASH
    └── Police Squad
```

Assets can (but on't have to) belong to one ore more collections, and a
collection of assets can be contained within another collection.
```
├── Movies
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
│       └── Star Trek Into Darkness (2013)
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

In this case, there is a Movies collection and a TV Shows collection at the top
level.  The Movies collection has a single Star Trek collection with all the
Star Trek movies contained with it.  Likewise, the House sub-collection is
divided into one sub-collection for each season of the show.

Users have access to the assets based on filters.  A user can only access the
assets that are exposed to him via filters applied to his profile.  A user with
no filters would have access to all assets/collections.  Filters can be applied
directly to a user or to a group and users can belong to zero or more groups.
Filters match on attributes of the assets' metadata as well as their folder and
collection memberships.
