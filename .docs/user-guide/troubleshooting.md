# Troubleshooting

---

## Log file location

ScaleIO-Framework keeps a log in the `stdout` file for the Marathon's task.

## Debug output

The ScaleIO-Framework can be started with the `-debug=true` flag to maximize log
detail.

To watch polly live, including error output and traces, you simply open the
`stdout` files on the Marathon task. A new window with the streaming output for
the `stdout` will be created.
