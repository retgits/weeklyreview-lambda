# Weekly Review

The weekly review uses cards from the Trello board **Main**, in the list **Done** to generate an overview of tasks.

Required environment variables are:

* TRELLO_APPKEY
* TRELLO_APPTOKEN
* TRELLO_MEMBERNAME

It produces a structure like below using labels that should match the headings (cards with the label `Planned Done` will show up under that heading).

```markdown
Weekly Review
# Planned Done
These are items / tasks that were planned last week, and executed exactly as planned. Good job! Lets reward somebody and sing some praises!
* Task 1
* Task 2

# Planned / Not Done
These are items / tasks that were planned for this week, but were not executed. The questions to asks are: What happened? Why? Who was responsible? What are the next steps?
* Task 1
* Task 2

# Unplanned / Done
These are items / tasks that were not planned, but were finished. This is not necessarily a good thing! Why was this worked on if it wasn’t planned? Who’s responsible?
* Task 1
* Task 2

Other tasks
* Task 1
* Task 2

# What went well

# What didn't go well

# Where do I need help

```

## License

See the [LICENSE](./LICENSE) file for details