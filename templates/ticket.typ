// SiteProof T&M Ticket Template
// Typst template for generating immutable PDF records

#let ticket(
  ticket_number: "",
  ticket_type: "Time & Materials",
  work_date: "",
  project_name: "",
  project_address: "",
  gc_company: "",
  sub_company: "",
  title: "",
  description: "",
  line_items: (),
  photos: (),
  foreman_signature: none,
  foreman_name: "",
  foreman_signed_at: "",
  gc_signature: none,
  gc_name: "",
  gc_signed_at: "",
  content_hash: "",
  generated_at: "",
) = {
  // Page setup
  set page(
    paper: "us-letter",
    margin: (top: 1in, bottom: 1in, left: 0.75in, right: 0.75in),
    footer: [
      #set text(size: 8pt, fill: gray)
      #grid(
        columns: (1fr, 1fr),
        align(left)[Document ID: #ticket_number],
        align(right)[Page #counter(page).display() of #locate(loc => counter(page).final(loc).first())],
      )
    ],
  )
  
  set text(font: "Arial", size: 10pt)
  
  // Header
  align(center)[
    #text(size: 18pt, weight: "bold")[#ticket_type Ticket]
    #v(0.25em)
    #text(size: 14pt, fill: gray)[#ticket_number]
  ]
  
  v(1em)
  
  // Project info box
  rect(
    width: 100%,
    inset: 12pt,
    stroke: 1pt + gray,
    radius: 4pt,
  )[
    #grid(
      columns: (1fr, 1fr),
      gutter: 1em,
      [
        *Project:* #project_name \
        *Address:* #project_address \
        *GC:* #gc_company
      ],
      [
        *Subcontractor:* #sub_company \
        *Work Date:* #work_date \
        *Status:* Approved & Locked
      ],
    )
  ]
  
  v(1em)
  
  // Title and Description
  text(size: 12pt, weight: "bold")[Description of Extra Work]
  v(0.5em)
  text(weight: "bold")[#title]
  v(0.25em)
  text[#description]
  
  v(1em)
  
  // Line items table
  if line_items.len() > 0 {
    text(size: 12pt, weight: "bold")[Line Items]
    v(0.5em)
    
    table(
      columns: (auto, 1fr, auto, auto, auto, auto),
      inset: 8pt,
      align: (left, left, right, left, right, right),
      stroke: 0.5pt + gray,
      
      // Header
      table.header(
        [*Type*], [*Description*], [*Qty*], [*Unit*], [*Rate*], [*Total*],
      ),
      
      // Data rows
      ..line_items.map(item => (
        item.at("type", default: ""),
        item.at("description", default: ""),
        str(item.at("quantity", default: 0)),
        item.at("unit", default: ""),
        if item.at("rate", default: none) != none { "$" + str(item.rate) } else { "-" },
        if item.at("total", default: none) != none { "$" + str(item.total) } else { "-" },
      )).flatten(),
    )
    
    v(1em)
  }
  
  // Photos section
  if photos.len() > 0 {
    text(size: 12pt, weight: "bold")[Photo Evidence]
    v(0.5em)
    
    grid(
      columns: (1fr, 1fr),
      gutter: 1em,
      ..photos.map(photo => {
        rect(
          width: 100%,
          inset: 8pt,
          stroke: 0.5pt + gray,
          radius: 4pt,
        )[
          #align(center)[
            #if photo.at("path", default: none) != none {
              image(photo.path, width: 100%)
            } else {
              rect(width: 100%, height: 150pt, fill: luma(240))[
                #align(center + horizon)[Photo: #photo.at("id", default: "")]
              ]
            }
          ]
          #v(0.25em)
          #text(size: 8pt)[
            #photo.at("caption", default: "No caption") \
            #if photo.at("taken_at", default: none) != none {
              [Taken: #photo.taken_at]
            }
          ]
        ]
      }),
    )
    
    v(1em)
  }
  
  // Signatures
  pagebreak(weak: true)
  
  text(size: 12pt, weight: "bold")[Signatures]
  v(0.5em)
  
  grid(
    columns: (1fr, 1fr),
    gutter: 2em,
    
    // Foreman signature
    rect(
      width: 100%,
      inset: 12pt,
      stroke: 1pt + gray,
      radius: 4pt,
    )[
      *Subcontractor Representative*
      #v(0.5em)
      #if foreman_signature != none {
        image(foreman_signature, height: 60pt)
      } else {
        rect(width: 100%, height: 60pt, stroke: 0.5pt + gray)[]
      }
      #v(0.25em)
      #line(length: 100%, stroke: 0.5pt)
      #text(size: 9pt)[
        Name: #foreman_name \
        Signed: #foreman_signed_at
      ]
    ],
    
    // GC signature
    rect(
      width: 100%,
      inset: 12pt,
      stroke: 1pt + gray,
      radius: 4pt,
    )[
      *GC Superintendent*
      #v(0.5em)
      #if gc_signature != none {
        image(gc_signature, height: 60pt)
      } else {
        rect(width: 100%, height: 60pt, stroke: 0.5pt + gray)[]
      }
      #v(0.25em)
      #line(length: 100%, stroke: 0.5pt)
      #text(size: 9pt)[
        Name: #gc_name \
        Signed: #gc_signed_at
      ]
    ],
  )
  
  v(2em)
  
  // Verification footer
  rect(
    width: 100%,
    inset: 12pt,
    fill: luma(245),
    stroke: 0.5pt + gray,
    radius: 4pt,
  )[
    #set text(size: 8pt, fill: gray)
    *Document Verification*
    
    This document has been digitally signed and locked. Any modification to this document
    after the lock date will invalidate the content hash below.
    
    #v(0.5em)
    #grid(
      columns: (auto, 1fr),
      gutter: 1em,
      [Content Hash:], [#raw(content_hash)],
      [Generated:], [#generated_at],
      [System:], [SiteProof - siteproof.app],
    )
  ]
}

// Example usage (commented out for template)
// #ticket(
//   ticket_number: "TM-2024-0042",
//   ticket_type: "Time & Materials",
//   work_date: "December 29, 2024",
//   project_name: "Downtown Office Tower",
//   project_address: "123 Main St, New York, NY 10001",
//   gc_company: "BuildCo General Contractors",
//   sub_company: "Acme Electrical",
//   title: "Extra conduit run - Building B",
//   description: "Ran 50 linear feet of additional 3/4\" EMT conduit from electrical room to server closet due to design change RFI-042.",
//   line_items: (
//     (type: "Labor", description: "Electrician - overtime", quantity: 4, unit: "hours", rate: 85, total: 340),
//     (type: "Material", description: "3/4\" EMT conduit", quantity: 50, unit: "linear ft", rate: 2.50, total: 125),
//     (type: "Equipment", description: "Scissor lift rental", quantity: 0.5, unit: "day", rate: 200, total: 100),
//   ),
//   photos: (),
//   foreman_name: "John Smith",
//   foreman_signed_at: "December 29, 2024 at 2:30 PM EST",
//   gc_name: "Mike Johnson",
//   gc_signed_at: "December 29, 2024 at 2:45 PM EST",
//   content_hash: "a1b2c3d4e5f6...",
//   generated_at: "December 29, 2024 at 2:46 PM EST",
// )

