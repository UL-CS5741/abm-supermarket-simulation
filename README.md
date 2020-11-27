## abm-supermarket-simulation

### Points to take Note of:
- Number of items in trolley to be generated and then it should be going to the checkout depending on the number of items
- Producer produces Customers and Consumer process the checkout
- Lost customers are those who are suppossed to wait in the queue where already 6 customers are ahead.
- There should be buffered channels for executing checkout (ie queue), so each queue will take care of each item counter.